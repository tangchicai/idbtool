package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type MatchItem struct {
	Key   string
	Value string
	Score int
}

type dynamicComparer struct {
	name string
}

func (c *dynamicComparer) Name() string {
	return c.name
}

func (c *dynamicComparer) Compare(a, b []byte) int {
	min := len(a)
	if len(b) < min {
		min = len(b)
	}
	for i := 0; i < min; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		return 0
	}
}

func (c *dynamicComparer) Separator(dst, a, b []byte) []byte {
	dst = append(dst[:0], a...)
	return dst
}

func (c *dynamicComparer) Successor(dst, b []byte) []byte {
	dst = append(dst[:0], b...)
	return dst
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
	// 先尝试一个默认 comparer
	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		Comparer: newDynamicComparer("default"),
	})
	if err == nil {
		return db, nil
	}

	// 提取 got comparer
	re := regexp.MustCompile(`got '([^']+)'`)
	m := re.FindStringSubmatch(err.Error())
	if len(m) < 2 {
		return nil, err
	}

	gotComparer := m[1]

	if !askRetry(gotComparer) {
		return nil, err
	}

	// 用提取出来的 comparer 名称重试
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
	searchText := flag.String("search", "", "search text")
	ignoreCase := flag.Bool("i", true, "ignore case")
	limit := flag.Int("limit", 0, "limit results, 0 means no limit")
	deleteMode := flag.Bool("delete", false, "delete matched items after search")
	flag.Parse()

	if *dbPath == "" {
		fmt.Println("missing -db")
		os.Exit(1)
	}
	if *searchText == "" {
		fmt.Println("missing -search")
		os.Exit(1)
	}

	db, err := openDBWithPrompt(*dbPath)
	if err != nil {
		fmt.Println("Open DB error:", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("DB opened successfully")

	matches, err := SearchDB(db, *searchText, *ignoreCase, *limit)
	if err != nil {
		fmt.Println("Search error:", err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Println("No matched items found.")
		return
	}

	fmt.Println()
	fmt.Println("Matched items:")
	for i, m := range matches {
		fmt.Printf("[%d] SCORE: %d\n", i, m.Score)
		fmt.Printf("    KEY: %s\n", m.Key)
		fmt.Printf("    VALUE: %s\n", m.Value)
		fmt.Println()
	}

	if !*deleteMode {
		return
	}

	fmt.Println("Delete options:")
	fmt.Println("  - Enter a single index to delete one item, e.g. 0")
	fmt.Println("  - Enter multiple indices separated by comma, e.g. 0,2,5")
	fmt.Println("  - Enter 'all' to delete all matched items")
	fmt.Print("Enter your choice: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("No input, cancelled.")
		return
	}

	if strings.EqualFold(input, "all") {
		for _, m := range matches {
			if err := db.Delete([]byte(m.Key), nil); err != nil {
				fmt.Println("Delete error on key:", m.Key, "err:", err)
				continue
			}
			fmt.Println("Deleted:", m.Key)
		}
		fmt.Println("Done.")
		return
	}

	parts := strings.Split(input, ",")
	deleted := 0

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		idx, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println("Invalid index:", p)
			continue
		}
		if idx < 0 || idx >= len(matches) {
			fmt.Println("Index out of range:", idx)
			continue
		}

		keyToDelete := matches[idx].Key
		if err := db.Delete([]byte(keyToDelete), nil); err != nil {
			fmt.Println("Delete error:", err)
			continue
		}
		fmt.Println("Deleted:", keyToDelete)
		deleted++
	}

	fmt.Printf("Done. Deleted %d item(s).\n", deleted)
}

func SearchDB(db *leveldb.DB, target string, ignoreCase bool, limit int) ([]MatchItem, error) {
	iter := db.NewIterator(&util.Range{}, nil)
	defer iter.Release()

	var results []MatchItem
	targetNorm := normalizeText(target)

	if ignoreCase {
		targetNorm = strings.ToLower(targetNorm)
	}

	for iter.Next() {
		key := string(iter.Key())
		valueBytes := iter.Value()
		value := string(valueBytes)

		score := 0

		keyNorm := normalizeText(key)
		valNorm := normalizeText(value)

		if ignoreCase {
			keyNorm = strings.ToLower(keyNorm)
			valNorm = strings.ToLower(valNorm)
		}

		if strings.Contains(keyNorm, targetNorm) {
			score += 10
		}
		if strings.Contains(valNorm, targetNorm) {
			score += 10
		}

		for _, s := range extractPrintableStrings(valueBytes) {
			ns := normalizeText(s)
			if ignoreCase {
				ns = strings.ToLower(ns)
			}
			if strings.Contains(ns, targetNorm) {
				score += 1
			}
		}

		if matchLoose(value, target, ignoreCase) {
			score += 2
		}

		if score > 0 {
			results = append(results, MatchItem{
				Key:   key,
				Value: value,
				Score: score,
			})

			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return results, nil
}

func matchLoose(s, target string, ignoreCase bool) bool {
	if ignoreCase {
		s = strings.ToLower(s)
		target = strings.ToLower(target)
	}
	return strings.Contains(s, target)
}

func normalizeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractPrintableStrings(b []byte) []string {
	var result []string
	var buf bytes.Buffer

	flush := func() {
		if buf.Len() >= 4 {
			result = append(result, buf.String())
		}
		buf.Reset()
	}

	for _, c := range b {
		if unicode.IsPrint(rune(c)) && c != 0 {
			buf.WriteByte(c)
		} else {
			flush()
		}
	}
	flush()

	return result
}
