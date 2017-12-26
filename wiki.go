package main

import (
	"html/template"
	"net/http"
	"os"
	"regexp"

	"log"

	"strings"

	"time"

	"strconv"

	"github.com/justinas/alice"       // Middleware chaining
	"github.com/russross/blackfriday" // Markdown lib
)

var wikiDir string
var tagDir string
var pubDir string
var ekey []byte

var encryptionFlag = []byte("ENCRYPTED")

var specialDir []string

type basePage struct {
	Title string
	Nav   nav
}
type wikiPage struct {
	Body      template.HTML
	Tags      string
	TagArray  []string
	Created   string
	Modified  string
	Published bool
	Encrypted bool
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
func getWikiPubFilename(name string) string {
	return pubDir + name
}
func (p *wikiPage) save(s storage) error {
	var err error
	filename := getWikiFilename(wikiDir, p.Title)
	body := []byte(p.Body)
	if p.Encrypted {
		body, err = encrypt(body, ekey)
		if err != nil {
			return err
		}
		body = append(encryptionFlag, body...)
	}
	err = s.storeFile(filename, body)
	if err != nil {
		return err
	}

	tagsfile := getWikiTagsFilename(p.Title)
	err = s.storeFile(tagsfile, []byte(p.Tags))
	if err != nil {
		return err
	}

	log.Printf("Pub flag %v\n", p.Published)
	if p.Published {
		pubfile := getWikiPubFilename(p.Title)
		log.Printf("Saving %v\n", pubfile)
		err = s.storeFile(pubfile, nil)
		if err != nil {
			return err
		}

	} else {
		// Need to delete the pub file if it exists
	}

	return nil
}

const (
	myHTMLFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

	myExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS |
		blackfriday.EXTENSION_FOOTNOTES
)

func convertMarkdown(page *wikiPage, err error) (*wikiPage, error) {
	if err != nil {
		return page, err
	}
	mdRender := blackfriday.HtmlRenderer(myHTMLFlags, "", "")
	page.Body = template.HTML(blackfriday.Markdown([]byte(page.Body), mdRender, myExtensions))
	return page, nil

}
func viewHandler(w http.ResponseWriter, r *http.Request, p *wikiPage, s storage) {
	p, err := convertMarkdown(s.getPage(p))
	if err != nil {
		p, err = s.checkForPDF(p)
		if err != nil {
			http.Redirect(w, r, "/wiki/edit/"+p.Title, http.StatusFound)
			return
		}
	} else {
		p.Body = template.HTML(parseWikiWords([]byte(p.Body)))
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, p *wikiPage, s storage) {
	p, _ = s.getPage(p)
	renderTemplate(w, "edit", p)
}

func makeSearchHandler(fn navFunc, s storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("term") // Get the search term
		if len(term) == 0 {
			http.NotFound(w, r)
			return
		}

		results := ParseQueryResults(s.searchPages(wikiDir, term))
		p := &searchPage{Results: results, basePage: basePage{Title: "Search", Nav: fn(s)}}

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

func simpleHandler(page string, fn navFunc, s storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, page, fn(s))
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request, wiki string, s storage) string {
	body := r.FormValue("body")
	log.Printf("Checkbox is : %v", r.FormValue("wikipub"))
	p := wikiPage{basePage: basePage{Title: wiki}, Body: template.HTML(body), Tags: r.FormValue("wikitags")}
	if r.FormValue("wikipub") == "on" {
		p.Published = true
	}
	if r.FormValue("wikicrypt") == "on" {
		p.Encrypted = true
	}

	err := p.save(s)
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
	"views/viewjs.html",
	"views/pub.html",
	"views/pubhome.html",
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, *wikiPage, storage), navfn navFunc, s storage) http.HandlerFunc {
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
		p := &wikiPage{basePage: basePage{Title: wword, Nav: navfn(s)}}
		fn(w, r, p, s)
	}
}

func processSave(fn func(http.ResponseWriter, *http.Request, string, storage) string, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2], s)

	}
}

func parseWikiWords(target []byte) []byte {
	var wikiWord = regexp.MustCompile(`\{\{([^\}^#]+)[#]*(.*)\}\}`)

	return wikiWord.ReplaceAll(target, []byte("<a href=\"/wiki/view/$1#$2\">$1</a>"))
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
	specialDir = []string{"tags", "pub"}
	config, err := LoadConfig()
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
	if !strings.HasSuffix(wikiDir, "/") {
		wikiDir = wikiDir + "/"
	}
	tagDir = wikiDir + "tags/" // Make sure this doesnt double up the / in the path...
	pubDir = wikiDir + "pub/"  // Make sure this doesnt double up the / in the path...

	ekey = []byte(config.EncryptionKey)

	os.Mkdir(config.WikiDir, 0755)
	os.Mkdir(config.WikiDir+"tags", 0755)

	authHandlers := alice.New(loggingHandler, auth.validate)
	noauthHandlers := alice.New(loggingHandler)

	httpmux := http.NewServeMux()
	httpsmux := httpmux
	// setup wiki on https
	if config.UseHTTPS {
		httpsmux = http.NewServeMux()
	}

	fstore := fileStorage{tagDir}

	httpsmux.Handle("/wiki", authHandlers.ThenFunc(simpleHandler("home", getNav, fstore)))
	httpsmux.Handle("/wiki/login/", noauthHandlers.ThenFunc(auth.loginHandler))
	httpsmux.Handle("/wiki/register/", noauthHandlers.ThenFunc(auth.registerHandler))
	httpsmux.Handle("/wiki/logout/", authHandlers.ThenFunc(logoutHandler))
	httpsmux.Handle("/wiki/search/", authHandlers.ThenFunc(makeSearchHandler(getNav, fstore)))
	httpsmux.Handle("/wiki/view/", authHandlers.ThenFunc(makeHandler(viewHandler, getNav, fstore)))
	httpsmux.Handle("/wiki/viewjs/", authHandlers.ThenFunc(simpleHandler("viewjs", getNav, fstore)))
	httpsmux.Handle("/wiki/edit/", authHandlers.ThenFunc(makeHandler(editHandler, getNav, fstore)))
	httpsmux.Handle("/wiki/save/", authHandlers.ThenFunc(processSave(saveHandler, fstore)))
	httpsmux.Handle("/wiki/raw/", http.StripPrefix("/wiki/raw/", http.FileServer(http.Dir(wikiDir))))
	httpsmux.Handle("/pub/", noauthHandlers.ThenFunc(makePubHandler(pubHandler, getNav, fstore)))
	httpsmux.Handle("/pub", noauthHandlers.ThenFunc(simpleHandler("pubhome", getPubNav, fstore)))
	httpsmux.Handle("/api", noauthHandlers.ThenFunc(apiHandler(innerAPIHandler, fstore)))

	if config.UseHTTPS {
		// Any routes that duplicate the http routing are only done here
		httpsmux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		go http.ListenAndServeTLS(
			":"+strconv.Itoa(config.HTTPSPort),
			config.CertPath,
			config.KeyPath,
			httpsmux)

		httpmux.HandleFunc("/wiki", redirectHandler(*config))
		httpsmux.Handle("/", http.FileServer(http.Dir("wwwroot")))
	} else {
	}

	// Listen for normal traffic against root
	httpmux.Handle("/", http.FileServer(http.Dir("wwwroot")))
	httpmux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	err = http.ListenAndServe(":"+strconv.Itoa(config.HTTPPort), httpmux)
	checkErr(err)

}
