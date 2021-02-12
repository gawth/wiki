package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var ParseData = []struct {
	source string
	res    QueryResults
}{
	{"wiki\t12\tsometext", QueryResults{"wiki", "12", "sometext"}},
	{"wiki\t12\tsometext#123", QueryResults{"wiki", "12", "sometext#123"}},
	{"wiki\t12\tsometext\tfred", QueryResults{"wiki", "12", "sometext\tfred"}},
	{"wiki\t12", QueryResults{"wiki", "12", ""}},
	{"wiki", QueryResults{"ERROR", "", "Invalid query result"}},
}

func TestParseQueryResults(t *testing.T) {
	for _, td := range ParseData {
		res := ParseQueryResults([]string{td.source})
		if res[0].WikiName != td.res.WikiName {
			t.Errorf("ParseQueryResult: Failed to extract wiki name %v: %v", res[0].WikiName, td.res.WikiName)
		}
		if res[0].LineNum != td.res.LineNum {
			t.Errorf("ParseQueryResult: Failed to extract line num %v: %v", res[0].LineNum, td.res.LineNum)
		}
		if res[0].Text != td.res.Text {
			t.Errorf("ParseQueryResult: Failed to extract text %v: %v", res[0].Text, td.res.Text)
		}

	}
}

func TestViewHandler(t *testing.T) {
	testStr := "This is a test"
	p := wikiPage{
		Body: template.HTML(testStr),
	}
	s := stubStorage{
		page: p,
		getPageFunc: func(pg *wikiPage) (*wikiPage, error) {
			return pg, nil
		},
	}

	req := httptest.NewRequest("GET", "http://localhost/wiki/view/test", nil)
	w := httptest.NewRecorder()

	viewHandler(w, req, &p, &s)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
	}
	if !strings.Contains(string(body), testStr) {
		t.Errorf("Failed to get %v in %v", testStr, string(body))
	}
}

func TestViewRedirect(t *testing.T) {
	p := wikiPage{basePage: basePage{Title: "Test Title"}}
	s := stubStorage{
		page:        p,
		expectederr: errors.New("Page not found"),
		getPageFunc: func(pg *wikiPage) (*wikiPage, error) {
			return nil, errors.New("Page not found")
		},
	}

	req := httptest.NewRequest("GET", "http://localhost/wiki/view/test", nil)
	w := httptest.NewRecorder()

	viewHandler(w, req, &p, &s)

	resp := w.Result()
	if resp.StatusCode != 302 {
		t.Errorf("No redirect, expected 302 but got %v", resp.StatusCode)
	}
}

func stubNavFunc(s storage) nav {
	return nav{}
}

func TestSearchHandler(t *testing.T) {
	s := stubStorage{}
	req := httptest.NewRequest("GET", "http://localhost/wiki/search?term=test", nil)
	w := httptest.NewRecorder()

	handler := makeSearchHandler(stubNavFunc, &s)
	handler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
	}
}
func TestDeletehHandler(t *testing.T) {
	called := 0
	stubrec := func(f string) {
		if f != "deleteFile" {
			t.Fatalf("Expected deleteFile but got: %v", f)
		}
		called++
	}
	s := stubStorage{loggerFunc: stubrec}
	p := wikiPage{basePage: basePage{Title: "test"}}
	req := httptest.NewRequest("POST", "http://localhost/wiki/delete/test", nil)
	w := httptest.NewRecorder()

	deleteHandler(w, req, &p, &s)

	resp := w.Result()

	// When we get a delete we redirect to home page...
	if resp.StatusCode != 302 {
		t.Errorf("Failed to get a 302 response, got %v", resp.StatusCode)
	}
	// Expecting two calls to delete file - the wiki file and the tags file
	if called != 2 {
		t.Errorf("Expected delete to be called %v but was called %v", 2, called)
	}
}
func TestMovehHandler(t *testing.T) {
	called := 0
	stubrec := func(f string) {
		if f != "moveFile" {
			t.Fatalf("Expected moveFile but got: %v", f)
		}
		called++
	}
	s := stubStorage{loggerFunc: stubrec}
	p := wikiPage{basePage: basePage{Title: "test"}}
	form := url.Values{}
	form.Add("to", "newtest")
	req := httptest.NewRequest("POST", "http://localhost/wiki/move/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	moveHandler(w, req, &p, &s)

	resp := w.Result()

	// Should redirect to the new URL
	if resp.StatusCode != 302 {
		t.Errorf("Failed to get a 302 response, got %v", resp.StatusCode)
	}
	url, err := resp.Location()
	if err != nil {
		t.Errorf("Got error back from response URL check: %v", err)
	}
	if url.Path != "/wiki/view/newtest" {
		t.Errorf("Expected /wiki/view/newtest but got %v from 302", url.Path)
	}
	// Expecting two calls to - the wiki file and the tags file
	if called != 2 {
		t.Errorf("Expected storage  to be called %v but was called %v", 2, called)
	}
}

type mdc struct {
	calledWith string
}

func (m *mdc) ConvertURL(url string) (string, error) {
	m.calledWith = url
	return "converted", nil
}

func TestScrapeHandler(t *testing.T) {
	url := "fred"
	reader := strings.NewReader("target=test&url=" + url)
	req := httptest.NewRequest("POST", "http://localhost/wiki/scrape", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	stubConverter := mdc{}
	handler := makeScrapeHandler(scrapeHandler, &stubConverter)
	handler(w, req)

	resp := w.Result()

	expectedStatus := 302
	if resp.StatusCode != expectedStatus {
		t.Errorf("Failed to get a %v response, got %v", expectedStatus, resp.StatusCode)
	}
	expectedLoc := "/wiki/view/test"
	if resp.Header.Get("Location") != expectedLoc {
		t.Errorf("Redirect location not correct, expected %v but got %v", expectedLoc, resp.Header.Get("Location"))
	}
	if stubConverter.calledWith != url {
		t.Errorf("Stub call expect %v but got '%v'", url, stubConverter.calledWith)
	}
}
