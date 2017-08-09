package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type wikiNav struct {
	Name   string
	IsDir  bool
	SubNav []wikiNav
}
type nav struct {
	Pages []string
	Wikis []wikiNav
	Tags  TagIndex
}

type navFunc func(storage) nav

type byModTime []os.FileInfo

func (m byModTime) Len() int           { return len(m) }
func (m byModTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byModTime) Less(i, j int) bool { return m[i].ModTime().Before(m[j].ModTime()) }

func contains(target string, in []string) bool {
	for _, d := range in {
		if target == d {
			return true
		}
	}
	return false
}

func closureProcess(root string, names *[]wikiNav) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return err
		}
		if info.IsDir() && contains(info.Name(), specialDir) {
			return filepath.SkipDir
		}
		// Ignore anything that isnt an md file
		if strings.HasSuffix(info.Name(), ".md") {
			tmp := wikiNav{
				Name: strings.Replace(strings.TrimSuffix(path, ".md"), root, "", 1),
			}
			*names = append(*names, tmp)
		}
		if strings.HasSuffix(info.Name(), ".pdf") {
			tmp := wikiNav{
				Name: strings.Replace(path, root, "", 1),
			}
			*names = append(*names, tmp)
		}
		if info.IsDir() && path != root {
			tmp := wikiNav{
				Name: strings.Replace(path, root, "", 1) + "/",
			}
			*names = append(*names, tmp)
		}
		return nil
	}

}
func walkWikiDir(path string) []wikiNav {
	var names []wikiNav

	err := filepath.Walk(path, closureProcess(path, &names))
	checkErr(err)

	return names

}

func getWikiList(root, path string) []wikiNav {
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	sort.Sort(sort.Reverse(byModTime(files)))

	var names []wikiNav
	for _, info := range files {
		if info.IsDir() && contains(info.Name(), specialDir) {
			continue
		}
		// Ignore anything that isnt an md file
		if strings.HasSuffix(info.Name(), ".md") {
			tmp := wikiNav{
				Name: strings.TrimSuffix(info.Name(), ".md"),
			}
			names = append(names, tmp)
		}
		if strings.HasSuffix(info.Name(), ".pdf") {
			tmp := wikiNav{
				Name: info.Name(),
			}
			names = append(names, tmp)
		}
		if info.IsDir() {
			tmp := wikiNav{
				Name:  info.Name() + "/",
				IsDir: true,
			}
			tmp.SubNav = getWikiList(root, path+"/"+info.Name())
			names = append(names, tmp)
		}
	}

	return names

}

func getNav(s storage) nav {
	return nav{
		Wikis: getWikiList(wikiDir, wikiDir),
		Tags:  s.IndexRawFiles(wikiDir, "PDF", s.IndexTags(tagDir)),
	}
}
