package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var (
	wikiDir        string
	tagDir         string
	pubDir         string
	ekey           []byte
	encryptionFlag = []byte("ENCRYPTED")
	specialDir     []string
)

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
	Index []string
}

type searchPage struct {
	basePage
	Results []QueryResults
}

type mdConverter interface {
	ConvertURL(string) (string, error)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
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
	filename := getWikiFilename(wikiDir, p.Title)
	body := []byte(p.Body)
	if p.Encrypted {
		var err error
		body, err = encrypt(body, ekey)
		if err != nil {
			return err
		}
		body = append(encryptionFlag, body...)
	}
	if err := s.storeFile(filename, body); err != nil {
		return err
	}

	tagsfile := getWikiTagsFilename(p.Title)
	if err := s.storeFile(tagsfile, []byte(p.Tags)); err != nil {
		return err
	}

	pubfile := getWikiPubFilename(p.Title)
	if p.Published {
		if err := s.storeFile(pubfile, nil); err != nil {
			return err
		}
	} else {
		if err := s.deleteFile(pubfile); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func convertMarkdown(page *wikiPage, err error) (*wikiPage, error) {
	if err != nil {
		return page, err
	}

	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")

	page.Body = template.HTML(regexp.MustCompile("\r\n").ReplaceAllString(string(page.Body), "\n"))

	unsafe := blackfriday.Run([]byte(page.Body),
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.HardLineBreak|
				blackfriday.HeadingIDs|
				blackfriday.AutoHeadingIDs,
		),
	)

	page.Body = template.HTML(p.SanitizeBytes(unsafe))
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

func makeSearchHandler(fn navFunc, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("term")
		if term == "" {
			http.NotFound(w, r)
			return
		}

		results := ParseQueryResults(s.searchPages(wikiDir, term))
		p := &searchPage{Results: results, basePage: basePage{Title: "Search", Nav: fn(s)}}

		renderTemplate(w, "search", p)
	}
}

func simpleHandler(page string, fn navFunc, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, page, fn(s))
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, wiki string, s storage) string {
	body := r.FormValue("body")
	body = regexp.MustCompile("\r\n").ReplaceAllString(body, "\n")

	p := wikiPage{basePage: basePage{Title: wiki}, Body: template.HTML(body), Tags: r.FormValue("wikitags")}
	if r.FormValue("wikipub") == "on" {
		p.Published = true
	}
	if r.FormValue("wikicrypt") == "on" {
		p.Encrypted = true
	}

	if err := p.save(s); err != nil {
		log.Printf("Error saving wiki page: %v", err) // Add logging here
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}
	http.Redirect(w, r, "/wiki/view/"+p.Title, http.StatusFound)

	return r.FormValue("wikitags")
}

func deleteHandler(w http.ResponseWriter, r *http.Request, p *wikiPage, s storage) {
	filename := getWikiFilename(wikiDir, p.Title)

	if err := s.deleteFile(filename); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tagsfile := getWikiTagsFilename(p.Title)
	if err := s.deleteFile(tagsfile); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/wiki", http.StatusFound)
}

func moveHandler(w http.ResponseWriter, r *http.Request, p *wikiPage, s storage) {
	from := getWikiFilename(wikiDir, p.Title)
	to := r.FormValue("to")
	if to == "" {
		http.Error(w, "Form param 'to' needs setting", http.StatusBadRequest)
		return
	}
	tofile := getWikiFilename(wikiDir, to)

	if err := s.moveFile(from, tofile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tagsfile := getWikiTagsFilename(p.Title)
	totags := getWikiTagsFilename(to)
	if err := s.moveFile(tagsfile, totags); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/wiki/view/"+to, http.StatusFound)
}

func scrapeHandler(w http.ResponseWriter, r *http.Request, mdc mdConverter, st storage) {
	url := r.FormValue("url")
	name := r.FormValue("target")

	body, err := mdc.ConvertURL(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO Pass in file store and then when convert is called save to a new file
	// Need a means of determining where to save the file to...perhaps whatever is
	// specified - that should work for folders, etc already :-)
	p := wikiPage{basePage: basePage{Title: name}, Body: template.HTML(body), Tags: "Scraped"}

	if err := p.save(st); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/wiki/view/"+name, http.StatusFound)
}

var templates = template.Must(template.ParseFiles(
	"views/edit.html",
	"views/view.html",
	"views/pub.html",
	"views/pubhome.html",
	"views/home.html",
	"views/list.html",
	"views/search.html",
	"views/index.html",
	"views/footer.html",
	"views/recents.html",
	"views/leftnav.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile(`^/wiki/(edit|save|view|search|delete|move|scrape)/([a-zA-Z0-9\.\-_ /]*)$`)

func makeHandler(fn func(http.ResponseWriter, *http.Request, *wikiPage, storage), navfn navFunc, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wword := r.URL.Query().Get("wword")
		if wword == "" {
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

func makeScrapeHandler(fn func(http.ResponseWriter, *http.Request, mdConverter, storage), mdc mdConverter, fs storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, mdc, fs)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), time.Since(start))
	})
}

func main() {
	specialDir = []string{"tags", "pub"}
	config, err := LoadConfig()
	checkErr(err)

	if config.Logfile != "" {
		f, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		checkErr(err)
		defer f.Close()
		log.SetOutput(f)
	}

	wikiDir = strings.TrimSuffix(config.WikiDir, "/") + "/"
	tagDir = wikiDir + "tags/"
	pubDir = wikiDir + "pub/"
	ekey = []byte(config.EncryptionKey)

	os.MkdirAll(tagDir, 0755)
	os.MkdirAll(pubDir, 0755)

	httpmux := http.NewServeMux()
	
	// Option 1: Using original cached storage
	cached := newCachedStorage(fileStorage{tagDir}, wikiDir, tagDir)
	fstore := &cached
	
	// Option 2: Using new configurable storage wrapped with caching (currently commented out)
	/*
	storageConfig := StorageConfig{
		WikiDir: wikiDir,
		TagDir:  tagDir,
		PubDir:  pubDir,
		EncKey:  ekey,
	}
	configStore := NewConfigurableStorage(storageConfig)
	// Wrap in cached storage to maintain caching functionality
	cached := newCachedStorage(configStore.fs, configStore.config.WikiDir, configStore.config.TagDir)
	fstore := &cached
	*/
	
	htmltomd := md.NewConverter("", true, nil)

	httpmux.Handle("/wiki", loggingHandler(simpleHandler("home", getNav, fstore)))
	httpmux.Handle("/wiki/list/", loggingHandler(simpleHandler("list", getNav, fstore)))
	httpmux.Handle("/wiki/search/", loggingHandler(makeSearchHandler(getNav, fstore)))
	httpmux.Handle("/wiki/view/", loggingHandler(makeHandler(viewHandler, getNav, fstore)))
	httpmux.Handle("/wiki/edit/", loggingHandler(makeHandler(editHandler, getNav, fstore)))
	httpmux.Handle("/wiki/save/", loggingHandler(processSave(saveHandler, fstore)))
	httpmux.Handle("/wiki/delete/", loggingHandler(makeHandler(deleteHandler, getNav, fstore)))
	httpmux.Handle("/wiki/move/", loggingHandler(makeHandler(moveHandler, getNav, fstore)))
	httpmux.Handle("/wiki/scrape/", loggingHandler(makeScrapeHandler(scrapeHandler, htmltomd, fstore)))
	httpmux.Handle("/wiki/raw/", http.StripPrefix("/wiki/raw/", http.FileServer(http.Dir(wikiDir))))
	httpmux.Handle("/pub/", loggingHandler(makePubHandler(pubHandler, getNav, fstore)))
	httpmux.Handle("/pub", loggingHandler(simpleHandler("pubhome", getPubNav, fstore)))
	httpmux.Handle("/api/", loggingHandler(apiHandler(innerAPIHandler, fstore)))
	httpmux.Handle("/api", loggingHandler(apiHandler(innerAPIHandler, fstore)))
	httpmux.Handle("/", http.FileServer(http.Dir("wwwroot")))
	httpmux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	checkErr(http.ListenAndServe(":"+strconv.Itoa(config.HTTPPort), httpmux))
}
