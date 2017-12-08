package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
