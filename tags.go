package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Tag used to store a tag and associated wiki titles
type Tag struct {
	TagName string
	Wikis   []string
}

// GetWikisForTag returns a list of wikis for a tag
func (t *Tag) GetWikisForTag() []string {
	return t.Wikis
}

// AddWiki adds a wiki title to the tag
func (t *Tag) AddWiki(wiki string) {
	t.Wikis = append(t.Wikis, wiki)
}

// GetTagsFromString takes a string of comma separated tags and converts to
// a slice of tag structs
func GetTagsFromString(tagstring string) []string {
	tagnames := strings.Split(tagstring, ",")

	return tagnames
}

// TagIndex holds a list of tag objects and allows adding of wiki data
type TagIndex map[string]Tag

// AssociateTagToWiki adds a wiki page to a tag in the index
func (t TagIndex) AssociateTagToWiki(wiki, tag string) {
	tag = strings.TrimSpace(tag)
	val, exists := t[tag]
	if !exists {
		val = Tag{TagName: tag}
	}
	val.AddWiki(wiki)
	t[tag] = val

}

// GetTag returns the Tag from the tag index
func (t TagIndex) GetTag(tag string) Tag {
	return t[tag]
}

// IndexTags reads tags files from the file system and constructs
// an index
func IndexTags(path string) TagIndex {
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
func IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {

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
