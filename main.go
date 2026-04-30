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

func openDBWithPrompt(dbPath string) (*leveldb.DB, error) {
	// 先用一个默认 comparer 尝试
	initialComparer := newDynamicComparer("default")
	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		Comparer: initialComparer,
	})
	if err == nil {
		return db, nil
	}

	// 从错误中提取 got comparer
	re := regexp.MustCompile(`got '([^']+)'`)
	m := re.FindStringSubmatch(err.Error())
	if len(m) < 2 {
		return nil, err
	}

	gotComparer := m[1]

	// 询问是否重试
	if !askRetry(gotComparer) {
		return nil, err
	}

	// 动态创建 comparer 再次打开
	db, err = leveldb.OpenFile(dbPath, &opt.Options{
		Comparer: newDynamicComparer(gotComparer),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	dbPath := flag.String("db", "", "leveldb path")
	search := flag.String("search", "", "search key")
	flag.Parse()

	if *dbPath == "" {
		fmt.Println("missing -db")
		os.Exit(1)
	}

	db, err := openDBWithPrompt(*dbPath)
	if err != nil {
		fmt.Println("Open DB error:", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("DB opened successfully")

	if *search != "" {
		val, err := db.Get([]byte(*search), nil)
		if err != nil {
			fmt.Println("search error:", err)
			return
		}
		fmt.Printf("value: %x\n", val)
	}
}
