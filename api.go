package main

import (
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
	if tag == "" {
		//json.NewEncoder(w).Encode(nil)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
