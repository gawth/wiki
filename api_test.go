package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApiHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "http://localhost/api?tag=fred", nil)
	w := httptest.NewRecorder()
	s := stubStorage{
		GetTagWikisFunc: func(tag string) Tag {
			return Tag{TagName: "fred"}
		},
	}

	innerAPIHandler(w, req, &s)

	resp := w.Result()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response,  error: %v", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
	}

	var results Tag
	err = json.Unmarshal(data, &results)
	if err != nil {
		t.Errorf("Failed to read json data, error: %v, data: '%v'", err.Error(), string(data))
	}
	if results.TagName != "fred" {
		t.Errorf("Got the wrong tag back, expected 'fred' but got '%v'", results.TagName)
	}
}

func TestApiHandlerNoTag(t *testing.T) {

	req := httptest.NewRequest("GET", "http://localhost/api", nil)
	w := httptest.NewRecorder()
	s := stubStorage{}

	innerAPIHandler(w, req, &s)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Failed to get a %v response, got %v", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestWikiApiGetHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "http://localhost/api?wiki=fred", nil)
	w := httptest.NewRecorder()
	s := stubStorage{
		getPageFunc: func(pg *wikiPage) (*wikiPage, error) {
			return pg, nil
		},
		GetTagWikisFunc: func(tag string) Tag {
			return Tag{TagName: "fred"}
		},
	}

	innerAPIHandler(w, req, &s)

	resp := w.Result()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response,  error: %v", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
	}

	var results wikiPage
	err = json.Unmarshal(data, &results)
	if err != nil {
		t.Errorf("Failed to read json data, error: %v, data: '%v'", err.Error(), string(data))
	}
	if results.Title != "fred" {
		t.Errorf("Got the wrong page back, expected 'fred' but got '%v'", results.Title)
	}
}
func TestWikiApiPostHandler(t *testing.T) {

	req := httptest.NewRequest("POST", "http://localhost/api?wiki=fred", strings.NewReader("Some markup"))
	w := httptest.NewRecorder()
	s := stubStorage{
		getPageFunc: func(pg *wikiPage) (*wikiPage, error) {
			return pg, nil
		},
		storeFileFunc: func(name string, body []byte) error {
			t.Logf("storeFile: Got %v/n", name)
			// Check the first part of the string as the store file func will be called for the tags file and the
			// md file
			if !strings.HasPrefix(name, "fred") {
				t.Errorf("expecting %v but got %v", "fred", name)
			}
			return nil
		},
	}

	innerAPIHandler(w, req, &s)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
	}
}
