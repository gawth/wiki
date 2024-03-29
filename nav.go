package main

import (
	"log"
	"sort"
	"time"
)

type wikiNav struct {
	Name    string
	URL     string
	ID      string
	IsDir   bool
	SubNav  []wikiNav
	Mod     time.Time
	ModStr  string
	Summary string
}
type nav struct {
	Pages   []string
	Wikis   []wikiNav
	Tags    TagIndex
	Recents []wikiNav
}

type navFunc func(storage) nav

type byModTime []wikiNav

func (m byModTime) Len() int           { return len(m) }
func (m byModTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byModTime) Less(i, j int) bool { return m[i].Mod.Before(m[j].Mod) }

func contains(target string, in []string) bool {
	for _, d := range in {
		if target == d {
			return true
		}
	}
	return false
}

func flattenWikis(current []wikiNav) []wikiNav {
	var newList []wikiNav

	for _, v := range current {
		if v.IsDir {
			newList = append(newList, flattenWikis(v.SubNav)...)
		} else {
			newList = append(newList, v)
		}
	}
	return newList
}
func genRecents(current []wikiNav) []wikiNav {
	var newList []wikiNav

	newList = flattenWikis(current)
	sort.Sort(sort.Reverse(byModTime(newList)))

	return newList

}
func getNav(s storage) nav {
	start := time.Now()
	wikis := s.IndexWikiFiles("", wikiDir)
	loadwikis := time.Now()
	tags := s.IndexTags(tagDir)
	loadtags := time.Now()
	indexedTags := s.IndexRawFiles(wikiDir, "PDF", tags)
	indexTags := time.Now()

	log.Printf("[nav] wikis %v", loadwikis.Sub(start))
	log.Printf("[nav] tags %v", loadtags.Sub(loadwikis))
	log.Printf("[nav] idxtags %v", indexTags.Sub(loadtags))
	return nav{
		Wikis:   wikis,
		Tags:    indexedTags,
		Recents: genRecents(wikis),
	}
}
