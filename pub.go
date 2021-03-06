package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var validPubPath = regexp.MustCompile("^/pub/([a-zA-Z0-9\\.\\-_ /]*)$")

func makePubHandler(fn func(http.ResponseWriter, *http.Request, *wikiPage, storage), navfn navFunc, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Path is : %v", r.URL.Path)
		m := validPubPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		title := m[1]
		p := &wikiPage{basePage: basePage{Title: title}}
		fn(w, r, p, s)
	}
}
func pubHandler(w http.ResponseWriter, r *http.Request, p *wikiPage, s storage) {
	p, err := convertMarkdown(s.getPage(p))
	if err != nil {
	} else {
		p.Body = template.HTML(parseWikiWords([]byte(p.Body)))
	}

	renderTemplate(w, "pub", p)
}
func getPubNav(s storage) nav {
	return nav{
		Pages: s.getPublicPages(),
	}
}
