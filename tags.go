package main

import "strings"

// Tag used to store a tag and associated wiki titles
type Tag struct {
	TagName string
	wikis   []string
}

// GetWikisForTag returns a list of wikis for a tag
func (t *Tag) GetWikisForTag(tag string) []string {
	return t.wikis
}

// AddWiki adds a wiki title to the tag
func (t *Tag) AddWiki(wiki string) {
	t.wikis = append(t.wikis, wiki)
}

// GetTags takes a string of comma separated tags and converts to
// a slice of tag structs
func GetTags(wiki, tagstring string) []Tag {
	tagnames := strings.Split(tagstring, ",")

	res := []Tag{}

	for _, t := range tagnames {
		res = append(res, Tag{TagName: t})
	}
	return res
}

// TagIndex holds a list of tag objects and allows adding of wiki data
type TagIndex map[string][]Tag

// AssociateTagWiki adds a wiki page to a tag in the index
func (t *TagIndex) AssociateTagWiki(wiki, tag string) {

}
