// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"

	"github.com/golang-commonmark/markdown"
)

const wikiDir = "wiki/"

type basePage struct {
	Title string
	Nav   []string
}
type wikiPage struct {
	Body     template.HTML
	Created  string
	Modified string
	basePage
}
type searchPage struct {
	basePage
	Results []QueryResults
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (p *wikiPage) save() error {
	filename := wikiDir + p.Title
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func convertMarkdown(page *wikiPage, err error) (*wikiPage, error) {
	if err != nil {
		return page, err
	}
	md := markdown.New(markdown.HTML(true))
	page.Body = template.HTML(md.RenderToString([]byte(page.Body)))
	return page, nil

}
func loadPage(p *wikiPage) (*wikiPage, error) {
	filename := wikiDir + p.Title

	file, err := os.Open(filename)
	if err != nil {
		return p, err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	p.Body = template.HTML(body)

	info, err := file.Stat()
	if err != nil {
		return p, err
	}

	p.Modified = info.ModTime().String()
	return p, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, p *wikiPage) {
	p, err := convertMarkdown(loadPage(p))
	if err != nil {
		http.Redirect(w, r, "/edit/"+p.Title, http.StatusFound)
		return
	}
	p.Body = template.HTML(parseWikiWords([]byte(p.Body)))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, p *wikiPage) {
	p, _ = loadPage(p)
	renderTemplate(w, "edit", p)
}

func searchHandler(fn navFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("term") // Get the search term
		if len(term) == 0 {
			http.NotFound(w, r)
			return
		}

		results := ParseQueryResults(SearchWikis(wikiDir, term))
		p := &searchPage{Results: results, basePage: basePage{Title: "Search", Nav: fn()}}

		renderTemplate(w, "search", p)
	}
}

type navFunc func() []string

func homeHandler(page string, fn navFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, page, fn())
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request, p *wikiPage) {
	body := r.FormValue("body")
	p.Body = template.HTML(body)
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles(
	"views/edit.html",
	"views/view.html",
	"views/home.html",
	"views/search.html",
	"views/index.html",
	"views/leftnav.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|search)/([a-zA-Z0-9 ]*)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, *wikiPage), navfn navFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wword := r.URL.Query().Get("wword") // Get the wiki word param if available
		if len(wword) == 0 {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return
			}
			wword = m[2]
		}
		p := &wikiPage{basePage: basePage{Title: wword, Nav: navfn()}}
		fn(w, r, p)
	}
}

type byModTime []os.FileInfo

func (m byModTime) Len() int           { return len(m) }
func (m byModTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byModTime) Less(i, j int) bool { return m[i].ModTime().Before(m[j].ModTime()) }

func getNav() []string {
	return getWikiList(wikiDir)
}
func getWikiList(path string) []string {
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	sort.Sort(sort.Reverse(byModTime(files)))

	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}

	return names

}

func parseWikiWords(target []byte) []byte {
	var wikiWord = regexp.MustCompile(`\{([^\}]+)\}`)

	return wikiWord.ReplaceAll(target, []byte("<a href=\"/view/$1\">$1</a>"))
}

func main() {

	os.Mkdir(wikiDir, 0755)
	http.HandleFunc("/", homeHandler("home", getNav))
	http.HandleFunc("/search/", searchHandler(getNav))
	http.HandleFunc("/view/", makeHandler(viewHandler, getNav))
	http.HandleFunc("/edit/", makeHandler(editHandler, getNav))
	http.HandleFunc("/save/", makeHandler(saveHandler, getNav))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8080", nil)
}
