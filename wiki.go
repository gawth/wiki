package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"

	"log"

	"strings"

	"time"

	"github.com/golang-commonmark/markdown"
	"github.com/justinas/alice"
)

var wikiDir string
var tagDir string

type basePage struct {
	Title string
	Nav   nav
}
type wikiPage struct {
	Body     template.HTML
	Tags     string
	TagArray []string
	Created  string
	Modified string
	basePage
}
type searchPage struct {
	basePage
	Results []QueryResults
}
type nav struct {
	Wikis []string
	Tags  TagIndex
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getWikiFilename(folder, name string) string {
	return folder + name + ".md"
}

func getWikiTagsFilename(name string) string {
	return tagDir + name
}
func (p *wikiPage) save() error {
	filename := getWikiFilename(wikiDir, p.Title)
	err := ioutil.WriteFile(filename, []byte(p.Body), 0600)
	if err != nil {
		return err
	}

	tagsfile := getWikiTagsFilename(p.Title)
	err = ioutil.WriteFile(tagsfile, []byte(p.Tags), 0600)
	if err != nil {
		return err
	}
	return nil
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
	filename := getWikiFilename(wikiDir, p.Title)

	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return p, err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return p, err
	}
	p.Body = template.HTML(body)

	info, err := file.Stat()
	if err != nil {
		log.Println(err)
		return p, err
	}

	p.Modified = info.ModTime().String()

	tags, err := ioutil.ReadFile(getWikiTagsFilename(p.Title))
	if err == nil {
		p.Tags = string(tags)
		p.TagArray = strings.Split(p.Tags, ",")
	}

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

type navFunc func() nav

func homeHandler(page string, fn navFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, page, fn())
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request, wiki string) string {
	body := r.FormValue("body")
	p := wikiPage{basePage: basePage{Title: wiki}, Body: template.HTML(body), Tags: r.FormValue("wikitags")}

	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}
	http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)

	return r.FormValue("wikitags")
}

var templates = template.Must(template.ParseFiles(
	"views/edit.html",
	"views/view.html",
	"views/login.html",
	"views/home.html",
	"views/search.html",
	"views/index.html",
	"views/footer.html",
	"views/leftnav.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|search)/([a-zA-Z0-9\\.\\-_ ]*)$")

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

func processSave(fn func(http.ResponseWriter, *http.Request, string) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])

	}
}

type byModTime []os.FileInfo

func (m byModTime) Len() int           { return len(m) }
func (m byModTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m byModTime) Less(i, j int) bool { return m[i].ModTime().Before(m[j].ModTime()) }

func getNav() nav {
	return nav{
		Wikis: getWikiList(wikiDir),
		Tags:  IndexTags(tagDir),
	}
}
func getWikiList(path string) []string {
	files, err := ioutil.ReadDir(path)
	checkErr(err)

	sort.Sort(sort.Reverse(byModTime(files)))

	var names []string
	for _, f := range files {
		// Ignore anything that isnt an md file
		if strings.HasSuffix(f.Name(), ".md") {
			names = append(names, strings.TrimSuffix(f.Name(), ".md"))
		}
	}

	return names

}

func parseWikiWords(target []byte) []byte {
	var wikiWord = regexp.MustCompile(`\{([^\}]+)\}`)

	return wikiWord.ReplaceAll(target, []byte("<a href=\"/view/$1\">$1</a>"))
}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func main() {
	f, err := os.OpenFile("wiki.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		checkErr(err)
	}
	defer f.Close()

	log.SetOutput(f)

	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.LoadCookieKey()

	auth := NewAuth(*config, persistUsers)

	wikiDir = config.WikiDir
	tagDir = wikiDir + "/tags/"

	os.Mkdir(config.WikiDir, 0755)
	os.Mkdir(config.WikiDir+"tags", 0755)

	authHandlers := alice.New(loggingHandler, auth.validate)
	noauthHandlers := alice.New(loggingHandler)

	http.Handle("/", authHandlers.ThenFunc(homeHandler("home", getNav)))
	http.Handle("/login/", noauthHandlers.ThenFunc(auth.loginHandler))
	http.Handle("/register/", noauthHandlers.ThenFunc(auth.registerHandler))
	http.Handle("/logout/", authHandlers.ThenFunc(logoutHandler))
	http.Handle("/search/", authHandlers.ThenFunc(searchHandler(getNav)))
	http.Handle("/view/", authHandlers.ThenFunc(makeHandler(viewHandler, getNav)))
	http.Handle("/edit/", authHandlers.ThenFunc(makeHandler(editHandler, getNav)))
	http.Handle("/save/", authHandlers.ThenFunc(processSave(saveHandler)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	err = http.ListenAndServeTLS(":443", "/home/gawth/ssl/server.crt", "/home/gawth/ssl/server.key", nil)
	checkErr(err)
}
