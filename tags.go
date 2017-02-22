package main

import "strings"

import "io/ioutil"

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
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	index := TagIndex(make(map[string]Tag))

	for _, f := range files {
		contents, err := ioutil.ReadFile(path + f.Name())
		checkErr(err)

		for _, t := range GetTagsFromString(string(contents)) {
			index.AssociateTagToWiki(f.Name(), t)
		}
	}
	return index
}

// IndexRawFiles adds in tags for a file extension tag
func IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	// Loop through the files, setup a tag for PDF (extension) and then add to TI
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), strings.ToLower(fileExtension)) {
			existing.AssociateTagToWiki(f.Name(), fileExtension)
		}
	}
	return existing

}
