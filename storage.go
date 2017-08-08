package main

import (
	"bytes"
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
