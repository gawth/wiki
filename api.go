package main

import (
	"encoding/json"
	"net/http"
)

func apiHandler(fn func(http.ResponseWriter, *http.Request, storage), s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, s)
	}
}

func innerAPIHandler(w http.ResponseWriter, r *http.Request, s storage) {
	w.Header().Set("Content-Type", "application/json")

	tag := r.URL.Query().Get("tag") // Get the tag
	// Just return an empty response if no tag found
	if tag != "" {
		data := s.GetTagWikis(tag)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
		return
	}
	wiki := r.URL.Query().Get("wiki") // Get the wiki
	if wiki != "" {
		wikipg := &wikiPage{basePage: basePage{Title: wiki}}
		wikipg, err := s.getPage(wikipg)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(wikipg)
			return
		}
		// dont insert code here unless you want to exe in the error case
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
