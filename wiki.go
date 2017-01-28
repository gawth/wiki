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

type Page struct {
	Title string
	Body  template.HTML
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (p *Page) save() error {
	filename := wikiDir + p.Title
	return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func convertMarkdown(page *Page, err error) (*Page, error) {
	if err != nil {
		return nil, err
	}
	md := markdown.New(markdown.HTML(true))
	page.Body = template.HTML(md.RenderToString([]byte(page.Body)))
	return page, nil

}
func loadPage(title string) (*Page, error) {
	filename := wikiDir + title
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := convertMarkdown(loadPage(title))
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

type getFiles func(string) []string

// Change this to take a func rather than a list so that it can refresh the files when required
// ...using a closure...I think :-)
func homeHandler(dir string, filesFunc getFiles) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "home", filesFunc(dir))
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("views/edit.html", "views/view.html", "views/home.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
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
		fn(w, r, wword)
	}
}

type byModTime []os.FileInfo

func (m byModTime) Len() int           { return len(m) }
func (m byModTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byModTime) Less(i, j int) bool { return m[i].ModTime().Before(m[j].ModTime()) }

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

func main() {

	os.Mkdir(wikiDir, 0755)
	http.HandleFunc("/", homeHandler(wikiDir, getWikiList))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil)
}
