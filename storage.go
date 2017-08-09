package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type storage interface {
	storeFile(name string, content []byte) error
	getPublicPages() []string
	getPage(p *wikiPage) (*wikiPage, error)
	searchPages(root, query string) []string
	checkForPDF(p *wikiPage) (*wikiPage, error)
	IndexTags(path string) TagIndex
	IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex
}

type fileStorage struct {
}

func createDir(filename string) error {
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs fileStorage) storeFile(name string, content []byte) error {
	err := createDir(name)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(name, content, 0600)
	if err != nil {
		return err
	}

	return nil
}

func indexPubPages(path string) []string {

	var results []string

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			results = append(results, strings.TrimPrefix(subpath, path))
		}
		return nil
	})
	checkErr(err)

	return results
}

func (fs fileStorage) getPublicPages() []string {
	return indexPubPages(pubDir)
}

func (fs fileStorage) getPage(p *wikiPage) (*wikiPage, error) {
	filename := getWikiFilename(wikiDir, p.Title)

	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return p, err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return p, err
	}
	if bytes.HasPrefix(body, encryptionFlag) {
		tmp := bytes.TrimPrefix(body, encryptionFlag)

		body, err = decrypt(tmp, ekey)
		if err != nil {
			log.Println(err)
			return p, err
		}
		p.Encrypted = true
	}
	p.Body = template.HTML(body)

	info, err := file.Stat()
	if err != nil {
		log.Println(err)
		return p, err
	}

	p.Modified = info.ModTime().String()

	tags, err := ioutil.ReadFile(getWikiTagsFilename(p.Title))
	if err == nil {
		p.Tags = string(tags)
		p.TagArray = strings.Split(p.Tags, ",")
	}

	pubfilename := getWikiPubFilename(p.Title)

	pubfile, err := os.Open(pubfilename)
	if err == nil {
		p.Published = true
		pubfile.Close()
	}

	return p, nil
}

func (fs fileStorage) searchPages(root string, query string) []string {
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

func (fs fileStorage) checkForPDF(p *wikiPage) (*wikiPage, error) {
	filename := getPDFFilename(wikiDir, p.Title)

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open %v, %v\n", p.Title, err.Error())
		return p, err
	}
	defer file.Close()

	p.Body = template.HTML(fmt.Sprintf("<a href=\"/wiki/raw/%v\">%v</a>", p.Title, p.Title))
	return p, nil
}

// IndexTags reads tags files from the file system and constructs
// an index
func (fs fileStorage) IndexTags(path string) TagIndex {
	index := TagIndex(make(map[string]Tag))

	log.Println("Tag base folder :" + path)

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, _ error) error {
		// log.Println("walk:" + subpath)
		if !info.IsDir() {
			contents, err := ioutil.ReadFile(subpath)
			checkErr(err)

			wikiName := strings.TrimPrefix(subpath, path)
			for _, t := range GetTagsFromString(string(contents)) {
				index.AssociateTagToWiki(wikiName, t)
			}
		}
		return nil
	})
	checkErr(err)

	return index
}

// IndexRawFiles adds in tags for a file extension tag
func (fs fileStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, _ error) error {
		if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(fileExtension)) {
			filename := strings.TrimPrefix(subpath, path)
			existing.AssociateTagToWiki(filename, fileExtension)
		}
		return nil
	})
	checkErr(err)

	return existing

}
