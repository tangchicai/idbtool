package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type dynamicComparer struct {
	name string
}

func (c *dynamicComparer) Name() string {
	return c.name
}

func (c *dynamicComparer) Compare(a, b []byte) int {
	return strings.Compare(string(a), string(b))
}

func (c *dynamicComparer) Separator(dst, a, b []byte) []byte {
	return a
}

func (c *dynamicComparer) Successor(dst, b []byte) []byte {
	return b
}

func newDynamicComparer(name string) comparer.Comparer {
	return &dynamicComparer{name: name}
}

func askRetry(got string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("comparer mismatch detected. got: %s\nRetry with this comparer? [Y/N]: ", got)

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToUpper(input))

		if input == "Y" {
			return true
		}
		if input == "N" {
			return false
		}
		fmt.Print("Please input Y or N: ")
	}
}

func extractGotComparer(err error) string {
	re := regexp.MustCompile(`got[: ]+([^\s]+)`)
	m := re.FindStringSubmatch(err.Error())
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func openDBWithAutoRetry(dbPath string) (*leveldb.DB, error) {
	// 第一次先用一个默认 comparer 打开
	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		Comparer: newDynamicComparer("default"),
	})
	if err == nil {
		return db, nil
	}

	got := extractGotComparer(err)
	if got == "" {
		return nil, err
	}

	if !askRetry(got) {
		return nil, err
	}

	// 使用 got comparer 再次打开
	db, err = leveldb.OpenFile(dbPath, &opt.Options{
		Comparer: newDynamicComparer(got),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func printKV(key, value []byte) {
	fmt.Printf("key=%q  value=%x\n", key, value)
}

func scanAll(db *leveldb.DB) error {
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		printKV(iter.Key(), iter.Value())
	}
	return iter.Error()
}

func scanPrefix(db *leveldb.DB, prefix string) error {
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	p := []byte(prefix)
	for iter.Seek(p); iter.Valid(); iter.Next() {
		k := iter.Key()
		if !strings.HasPrefix(string(k), prefix) {
			break
		}
		printKV(k, iter.Value())
	}
	return iter.Error()
}

func searchKey(db *leveldb.DB, key string) error {
	val, err := db.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	printKV([]byte(key), val)
	return nil
}

func main() {
	dbPath := flag.String("db", "", "leveldb path")
	search := flag.String("search", "", "search key")
	prefix := flag.String("prefix", "", "prefix search")
	scan := flag.Bool("scan", false, "scan all keys")
	flag.Parse()

	if *dbPath == "" {
		fmt.Println("missing -db")
		os.Exit(1)
	}

	db, err := openDBWithAutoRetry(*dbPath)
	if err != nil {
		fmt.Println("Open DB error:", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("DB opened successfully")

	switch {
	case *search != "":
		err := searchKey(db, *search)
		if err != nil {
			fmt.Println("search error:", err)
		}
	case *prefix != "":
		err := scanPrefix(db, *prefix)
		if err != nil {
			fmt.Println("prefix scan error:", err)
		}
	case *scan:
		err := scanAll(db)
		if err != nil {
			fmt.Println("scan error:", err)
		}
	default:
		fmt.Println("No action specified. Use -search, -prefix or -scan")
	}
}
