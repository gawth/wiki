package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"log"

	"fmt"
	"strings"

	"time"

	"strconv"

	"github.com/golang-commonmark/markdown"
	"github.com/justinas/alice"
)

var wikiDir string
var tagDir string

var specialDir []string

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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func getPDFFilename(folder, name string) string {
	return folder + name
}

func getWikiFilename(folder, name string) string {
	return folder + name + ".md"
}

func getWikiTagsFilename(name string) string {
	return tagDir + name
}
func createDir(filename string) error {
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *wikiPage) save() error {
	filename := getWikiFilename(wikiDir, p.Title)

	err := createDir(filename)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, []byte(p.Body), 0600)
	if err != nil {
		return err
	}

	tagsfile := getWikiTagsFilename(p.Title)

	err = createDir(tagsfile)
	if err != nil {
		return err
	}
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

func checkForPDF(p *wikiPage) (*wikiPage, error) {
	filename := getPDFFilename(wikiDir, p.Title)

	file, err := os.Open(filename)
	if err != nil {
		return p, err
	}
	defer file.Close()

	p.Body = template.HTML(fmt.Sprintf("<a href=\"/wiki/raw/%v\">%v</a>", p.Title, p.Title))
	return p, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, p *wikiPage) {
	p, err := convertMarkdown(loadPage(p))
	if err != nil {
		p, err = checkForPDF(p)
		if err != nil {
			http.Redirect(w, r, "/wiki/edit/"+p.Title, http.StatusFound)
			return
		}
	} else {
		p.Body = template.HTML(parseWikiWords([]byte(p.Body)))
	}

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

func redirectHandler(c Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		var port string
		hostparts := strings.Split(host, ":")
		if len(hostparts) == 2 {
			host = hostparts[0]
			port = strconv.Itoa(c.HTTPSPort)
		}
		target := "https://" + host
		if len(port) > 0 {
			target += ":" + port

		}
		target += r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)

	}
}

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
	http.Redirect(w, r, "/wiki/view/"+p.Title, http.StatusFound)

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

var validPath = regexp.MustCompile("^/wiki/(edit|save|view|search)/([a-zA-Z0-9\\.\\-_ /]*)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, *wikiPage), navfn navFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wword := r.URL.Query().Get("wword") // Get the wiki word param if available
		if len(wword) == 0 {
			log.Printf("Path is : %v", r.URL.Path)
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

func parseWikiWords(target []byte) []byte {
	var wikiWord = regexp.MustCompile(`\{\{([^\}]+)\}\}`)

	return wikiWord.ReplaceAll(target, []byte("<a href=\"/wiki/view/$1\">$1</a>"))
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
	specialDir = []string{"tags"}
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	if config.Logfile != "" {
		f, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			checkErr(err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	config.LoadCookieKey()

	auth := NewAuth(*config, persistUsers)

	wikiDir = config.WikiDir
	tagDir = wikiDir + "/tags/"

	os.Mkdir(config.WikiDir, 0755)
	os.Mkdir(config.WikiDir+"tags", 0755)

	authHandlers := alice.New(loggingHandler, auth.validate)
	noauthHandlers := alice.New(loggingHandler)

	// Listen for normal traffic against root
	httpmux := http.NewServeMux()
	httpmux.Handle("/", http.FileServer(http.Dir("wwwroot")))
	httpmux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	httpmux.HandleFunc("/wiki", redirectHandler(*config))
	go http.ListenAndServe(":"+strconv.Itoa(config.HTTPPort), httpmux)

	// setup wiki on https
	httpsmux := http.NewServeMux()
	httpsmux.Handle("/wiki", authHandlers.ThenFunc(homeHandler("home", getNav)))
	httpsmux.Handle("/wiki/login/", noauthHandlers.ThenFunc(auth.loginHandler))
	httpsmux.Handle("/wiki/register/", noauthHandlers.ThenFunc(auth.registerHandler))
	httpsmux.Handle("/wiki/logout/", authHandlers.ThenFunc(logoutHandler))
	httpsmux.Handle("/wiki/search/", authHandlers.ThenFunc(searchHandler(getNav)))
	httpsmux.Handle("/wiki/view/", authHandlers.ThenFunc(makeHandler(viewHandler, getNav)))
	httpsmux.Handle("/wiki/edit/", authHandlers.ThenFunc(makeHandler(editHandler, getNav)))
	httpsmux.Handle("/wiki/save/", authHandlers.ThenFunc(processSave(saveHandler)))
	httpsmux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	httpsmux.Handle("/wiki/raw/", http.StripPrefix("/wiki/raw/", http.FileServer(http.Dir(wikiDir))))

	err = http.ListenAndServeTLS(
		":"+strconv.Itoa(config.HTTPSPort),
		config.CertPath,
		config.KeyPath,
		httpsmux)
	checkErr(err)
}
