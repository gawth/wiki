package main

import "strings"

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
	// Calling split on empty seems to give you a slice with an empty string
	// so need to manually check before going ahead with the split and return a
	// empty slice
	if tagstring == "" {
		return []string{}
	}
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
