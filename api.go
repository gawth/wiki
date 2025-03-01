package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	if r.Method != "POST" {
		return false
	}

	wiki := r.URL.Query().Get("wiki") // Get the wiki
	if wiki == "" {
		return false
	}

	body, err := io.ReadAll(r.Body)
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

	w.WriteHeader(http.StatusOK)
	return true
}
func handleGetList(w http.ResponseWriter, r *http.Request, s storage) bool {
	if r.Method != "GET" {
		return false
	}

	list := r.URL.Query().Get("list") // Get the wiki list
	if list == "" {
		return false
	}
	data := s.getWikiList(list)

	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}

	return true
}

func handleImageUpload(w http.ResponseWriter, r *http.Request, s storage) bool {
	// Extract wiki title from URL path
	parts := strings.Split(r.URL.Path, "/")
	
	if len(parts) < 4 || parts[2] != "image" {
		return false
	}
	wikiTitle := parts[3]
	
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return true
	}
	
	// Get image file
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return true
	}
	defer file.Close()
	
	// Read file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	
	// Get file extension or default to .png
	fileExt := filepath.Ext(handler.Filename)
	if fileExt == "" {
		fileExt = ".png"
	}
	
	var imageURL string
	
	// Check if resize parameters were provided
	widthStr := r.FormValue("width")
	heightStr := r.FormValue("height")
	
	if widthStr != "" || heightStr != "" {
		// Parse resize dimensions
		width, height := 0, 0
		var parseErr error
		
		if widthStr != "" {
			width, parseErr = parseInt(widthStr)
			if parseErr != nil {
				log.Printf("Error parsing width parameter: %v", parseErr)
				// Default to a sensible width if parse fails
				width = 800
			}
		}
		
		if heightStr != "" {
			height, parseErr = parseInt(heightStr)
			if parseErr != nil {
				log.Printf("Error parsing height parameter: %v", parseErr)
				// Default to a sensible height if parse fails
				height = 600
			}
		}
		
		// Ensure at least one dimension is specified
		if width <= 0 && height <= 0 {
			width = 800 // Default width if both dimensions are invalid
		}
		
		log.Printf("Resizing image to %dx%d", width, height)
		
		// Store resized image
		imageURL, err = s.storeResizedImage(wikiTitle, imageData, fileExt, width, height)
		if err != nil {
			log.Printf("Error resizing image: %v", err)
		}
	} else {
		// Store original image
		imageURL, err = s.storeImage(wikiTitle, imageData, fileExt)
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	
	// Return URL to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": imageURL})
	
	return true
}

// Helper function to parse integer from string
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func innerAPIHandler(w http.ResponseWriter, r *http.Request, s storage) {
	w.Header().Set("Content-Type", "application/json")

	// Add this before other handlers
	if strings.HasPrefix(r.URL.Path, "/api/image/") {
		if handleImageUpload(w, r, s) {
			return
		}
	}

	if ok := handleTag(w, r, s); ok {
		return
	}

	if ok := handleGetWiki(w, r, s); ok {
		return
	}

	if ok := handlePostWiki(w, r, s); ok {
		return
	}
	if ok := handleGetList(w, r, s); ok {
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
