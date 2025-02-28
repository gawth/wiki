package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
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

// StorageConfig holds configuration for file storage
type StorageConfig struct {
	WikiDir string
	TagDir  string
	PubDir  string
	EncKey  []byte
}

// ConfigurableStorage is a storage implementation that does not rely on global variables
type ConfigurableStorage struct {
	config StorageConfig
	fs     fileStorage
}

// NewConfigurableStorage creates a new configurable storage with the given config
func NewConfigurableStorage(config StorageConfig) *ConfigurableStorage {
	return &ConfigurableStorage{
		config: config,
		fs:     fileStorage{TagDir: config.TagDir},
	}
}

// Implement the storage interface methods for ConfigurableStorage
func (cs *ConfigurableStorage) storeFile(name string, content []byte) error {
	return cs.fs.storeFile(name, content)
}

func (cs *ConfigurableStorage) deleteFile(name string) error {
	return cs.fs.deleteFile(name)
}

func (cs *ConfigurableStorage) moveFile(from, to string) error {
	return cs.fs.moveFile(from, to)
}

func (cs *ConfigurableStorage) getPublicPages() []string {
	// Replace global pubDir with config.PubDir
	originalPubDir := pubDir
	pubDir = cs.config.PubDir
	defer func() { pubDir = originalPubDir }()
	
	return cs.fs.getPublicPages()
}

func (cs *ConfigurableStorage) getPage(p *wikiPage) (*wikiPage, error) {
	// Replace globals with config values
	originalWikiDir := wikiDir
	originalTagDir := tagDir
	originalPubDir := pubDir
	originalEkey := ekey
	
	wikiDir = cs.config.WikiDir
	tagDir = cs.config.TagDir
	pubDir = cs.config.PubDir
	ekey = cs.config.EncKey
	
	defer func() {
		wikiDir = originalWikiDir
		tagDir = originalTagDir
		pubDir = originalPubDir
		ekey = originalEkey
	}()
	
	return cs.fs.getPage(p)
}

func (cs *ConfigurableStorage) searchPages(root, query string) []string {
	return cs.fs.searchPages(root, query)
}

func (cs *ConfigurableStorage) checkForPDF(p *wikiPage) (*wikiPage, error) {
	// Replace wikiDir with config.WikiDir
	originalWikiDir := wikiDir
	wikiDir = cs.config.WikiDir
	defer func() { wikiDir = originalWikiDir }()
	
	return cs.fs.checkForPDF(p)
}

func (cs *ConfigurableStorage) IndexTags(path string) TagIndex {
	return cs.fs.IndexTags(path)
}

func (cs *ConfigurableStorage) GetTagWikis(tag string) Tag {
	return cs.fs.GetTagWikis(tag)
}

func (cs *ConfigurableStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	return cs.fs.IndexRawFiles(path, fileExtension, existing)
}

func (cs *ConfigurableStorage) IndexWikiFiles(base, path string) []wikiNav {
	return cs.fs.IndexWikiFiles(base, path)
}

func (cs *ConfigurableStorage) getWikiList(from string) []string {
	// Replace wikiDir with config.WikiDir
	originalWikiDir := wikiDir
	wikiDir = cs.config.WikiDir
	defer func() { wikiDir = originalWikiDir }()
	
	return cs.fs.getWikiList(from)
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

	err = os.WriteFile(name, content, 0600)
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

	body, err := io.ReadAll(file)
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

	tags, err := os.ReadFile(getWikiTagsFilename(p.Title))
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
			contents, err := os.ReadFile(subpath)
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
	files, err := os.ReadDir(path)
	checkErr(err)

	var names []wikiNav
	for _, f := range files {

		if f.IsDir() && contains(f.Name(), specialDir) {
			continue
		}
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		info, err := f.Info()
		checkErr(err)

		// Ignore anything that isnt an md file
		if strings.HasSuffix(f.Name(), ".md") {
			name := strings.TrimSuffix(f.Name(), ".md")
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
		if strings.HasSuffix(f.Name(), ".txt") {
			name := strings.TrimSuffix(f.Name(), ".txt")
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
		if strings.HasSuffix(f.Name(), ".pdf") {
			tmp := wikiNav{
				Name:   f.Name(),
				URL:    base + "/" + f.Name(),
				Mod:    info.ModTime(),
				ModStr: info.ModTime().Format(TIME_FORMAT),
				ID:     genID(base, f.Name()),
			}
			names = append(names, tmp)
		}
		if f.IsDir() {
			newbase := base + "/" + info.Name()
			tmp := wikiNav{
				Name:  f.Name(),
				URL:   newbase,
				IsDir: true,
				ID:    genID(base, f.Name()),
			}
			tmp.SubNav = fst.IndexWikiFiles(newbase, path+"/"+f.Name())
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
