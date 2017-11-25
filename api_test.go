package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiHandler(t *testing.T) {

	req := httptest.NewRequest("GET", "http://localhost/api?tag=fred", nil)
	w := httptest.NewRecorder()
	s := stubStorage{}

	innerAPIHandler(w, req, &s)

	resp := w.Result()
	ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("Failed to get a 200 response, got %v", resp.StatusCode)
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
