package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http/httptest"
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
	deletecalled := 0
	stubrec := func(f string) {
		deletecalled++
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
	if deletecalled != 2 {
		t.Errorf("Expected delete to be called %v but was called %v", 2, deletecalled)
	}
}
