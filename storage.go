package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const TIME_FORMAT = "2006-01-02 15:04:05"

type storage interface {
	storeFile(name string, content []byte) error
	deleteFile(name string) error
	moveFile(from, to string) error
	getPublicPages() []string
	getPage(p *wikiPage) (*wikiPage, error)
	searchPages(root, query string) []string
	checkForPDF(p *wikiPage) (*wikiPage, error)
	IndexTags(path string) TagIndex
	GetTagWikis(tag string) Tag
	IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex
	IndexWikiFiles(base, path string) []wikiNav
	getWikiList(from string) []string
}

type fileStorage struct {
	TagDir string
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
func (fst *fileStorage) storeFile(name string, content []byte) error {
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

func (fst *fileStorage) deleteFile(name string) error {
	if err := os.Remove(name); err != nil {
		return err
	}

	return nil
}
func (fst *fileStorage) moveFile(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return err
	}
	return nil
}

func indexPubPages(path string) []string {

	var results []string

	err := filepath.WalkDir(path, func(subpath string, info fs.DirEntry, _ error) error {
		if !info.IsDir() {
			results = append(results, strings.TrimPrefix(subpath, path))
		}
		return nil
	})
	checkErr(err)

	return results
}

func (fst *fileStorage) getPublicPages() []string {
	return indexPubPages(pubDir)
}

func (fst *fileStorage) getPage(p *wikiPage) (*wikiPage, error) {
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

func (fst *fileStorage) searchPages(root string, query string) []string {
	var wg sync.WaitGroup
	results := make(chan string)

	filepath.WalkDir(root, func(path string, file fs.DirEntry, err error) error {
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
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		if strings.Contains(scanner.Text(), query) {
			match := fmt.Sprintf("%s\t%d\t%s\n", name, i, scanner.Text())
			results <- match
		}
	}
}

func (fst *fileStorage) checkForPDF(p *wikiPage) (*wikiPage, error) {
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
func (fst *fileStorage) IndexTags(path string) TagIndex {
	index := TagIndex(make(map[string]Tag))

	err := filepath.WalkDir(path, func(subpath string, info fs.DirEntry, _ error) error {
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
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
func (fst *fileStorage) GetTagWikis(tag string) Tag {
	ti := fst.IndexTags(fst.TagDir)
	return ti[tag]
}

// IndexRawFiles adds in tags for a file extension tag
func (fst *fileStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {

	err := filepath.WalkDir(path, func(subpath string, info fs.DirEntry, _ error) error {
		if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(fileExtension)) {
			filename := strings.TrimPrefix(subpath, path)
			existing.AssociateTagToWiki(filename, fileExtension)
		}
		return nil
	})
	checkErr(err)

	return existing

}

func genID(base, name string) string {
	return strings.ReplaceAll(base+name, "/", "-")
}

// IndexWikiFiles will crawl through picking out files that conform to requirements for wiki entries
// This includes md and pdf files.
// Any hidden (dot) files are skipped
// Folders are included as part of the path
// Mod time is used to order the files
func (fst *fileStorage) IndexWikiFiles(base, path string) []wikiNav {
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	var names []wikiNav
	for _, info := range files {

		if info.IsDir() && contains(info.Name(), specialDir) {
			continue
		}
		if strings.HasPrefix(info.Name(), ".") {
			continue
		}
		// Ignore anything that isnt an md file
		if strings.HasSuffix(info.Name(), ".md") {
			name := strings.TrimSuffix(info.Name(), ".md")
			tmp := wikiNav{
				Name:    name,
				URL:     base + "/" + name,
				Mod:     info.ModTime(),
				ModStr:  info.ModTime().Format(TIME_FORMAT),
				ID:      genID(base, name),
				Summary: "This is a test summary for markdown",
			}
			names = append(names, tmp)
		}
		if strings.HasSuffix(info.Name(), ".txt") {
			name := strings.TrimSuffix(info.Name(), ".txt")
			tmp := wikiNav{
				Name:    name,
				URL:     base + "/" + name,
				Mod:     info.ModTime(),
				ModStr:  info.ModTime().Format(TIME_FORMAT),
				ID:      genID(base, name),
				Summary: "This is a test summary for text file",
			}
			names = append(names, tmp)
		}
		if strings.HasSuffix(info.Name(), ".pdf") {
			tmp := wikiNav{
				Name:   info.Name(),
				URL:    base + "/" + info.Name(),
				Mod:    info.ModTime(),
				ModStr: info.ModTime().Format(TIME_FORMAT),
				ID:     genID(base, info.Name()),
			}
			names = append(names, tmp)
		}
		if info.IsDir() {
			newbase := base + "/" + info.Name()
			tmp := wikiNav{
				Name:  info.Name(),
				URL:   newbase,
				IsDir: true,
				ID:    genID(base, info.Name()),
			}
			tmp.SubNav = fst.IndexWikiFiles(newbase, path+"/"+info.Name())
			if len(tmp.SubNav) > 0 {
				// Override the dir's mod time with the first entry
				tmp.Mod = tmp.SubNav[0].Mod
			}
			names = append(names, tmp)
		}
	}

	sort.Sort(sort.Reverse(byModTime(names)))
	return names

}

func (fst *fileStorage) getWikiList(from string) []string {
	path := wikiDir + from

	var results []string

	err := filepath.WalkDir(path, func(subpath string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Print("getWikiList", err)
			return nil
		}
		if !info.IsDir() {
			tmp := strings.TrimPrefix(subpath, wikiDir)
			tmp = strings.TrimSuffix(tmp, ".md")
			results = append(results, tmp)
		}
		return nil
	})
	checkErr(err)

	return results
}
