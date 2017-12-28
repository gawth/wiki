package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func apiHandler(fn func(http.ResponseWriter, *http.Request, storage), s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, s)
	}
}

func handleTag(w http.ResponseWriter, r *http.Request, s storage) bool {
	tag := r.URL.Query().Get("tag") // Get the tag
	// Just return an empty response if no tag found
	if tag == "" {
		return false
	}
	data := s.GetTagWikis(tag)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
	return true
}
func handleGetWiki(w http.ResponseWriter, r *http.Request, s storage) bool {
	if r.Method != "GET" {
		log.Println("Not a GET")
		return false
	}

	wiki := r.URL.Query().Get("wiki") // Get the wiki
	if wiki == "" {
		return false
	}
	wikipg := &wikiPage{basePage: basePage{Title: wiki}}
	wikipg, err := s.getPage(wikipg)
	if err != nil {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wikipg)
	return true
}
func handlePostWiki(w http.ResponseWriter, r *http.Request, s storage) bool {
	log.Println("Handling POST")
	if r.Method != "POST" {
		log.Println("Not a post")
		return false
	}

	wiki := r.URL.Query().Get("wiki") // Get the wiki
	if wiki == "" {
		log.Println("No wiki param")
		return false
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	var wp wikiPage
	if err := json.Unmarshal(body, &wp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	// TODO: Handle encryption and published pages

	err = wp.save(s)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	log.Printf("Saved %v\n", body)

	w.WriteHeader(http.StatusOK)
	return true
}

func innerAPIHandler(w http.ResponseWriter, r *http.Request, s storage) {
	w.Header().Set("Content-Type", "application/json")

	if ok := handleTag(w, r, s); ok {
		return
	}

	if ok := handleGetWiki(w, r, s); ok {
		return
	}

	if ok := handlePostWiki(w, r, s); ok {
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
