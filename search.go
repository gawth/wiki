package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func readFile(wg *sync.WaitGroup, name string, path string, query string, results chan string) {
	defer wg.Done()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		if strings.Contains(scanner.Text(), query) {
			match := fmt.Sprintf("%s\t%d\t%s\n", name, i, scanner.Text())
			results <- match
		}
	}
}

// SearchWikis looks through wiki files for some text
func SearchWikis(root string, query string) []string {
	var wg sync.WaitGroup
	results := make(chan string)

	filepath.Walk(root, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			wg.Add(1)
			name := strings.TrimSuffix(strings.TrimPrefix(path, root), ".md")
			go readFile(&wg, name, path, query, results)
		}
		return nil
	})
	go func() {
		wg.Wait()
		close(results)
	}()

	hits := []string{}
	for res := range results {
		hits = append(hits, res)
	}
	return hits
}

// QueryResults is used to hold search results after a wiki search
type QueryResults struct {
	WikiName string
	LineNum  string
	Text     string
}

// ParseQueryResults converts a result string to a query result
func ParseQueryResults(source []string) []QueryResults {
	res := []QueryResults{}
	for _, r := range source {
		sub := strings.Split(r, "\t")
		if len(sub) < 2 {
			res = append(res, QueryResults{"ERROR", "", "Invalid query result"})
			continue
		}
		res = append(res, QueryResults{
			WikiName: sub[0],
			LineNum:  sub[1],
			Text:     strings.Join(sub[2:], "\t"),
		})

	}
	return res
}
