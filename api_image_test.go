package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockImageStorage for testing image uploads
type MockImageStorage struct {
	storage
	imageURL string
	err      error
}

func (m *MockImageStorage) storeImage(wikiTitle string, imageData []byte, extension string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.imageURL, nil
}

func TestHandleImageUpload(t *testing.T) {
	// Setup mock storage
	mockURL := "/wiki/raw/images/test/12345.png"
	mockStorage := &MockImageStorage{imageURL: mockURL}
	
	// Create a test multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add image file to form
	fw, err := w.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	imageData := []byte("fake image data")
	fw.Write(imageData)
	w.Close()
	
	// Create test request
	req := httptest.NewRequest("POST", "/api/image/testpage", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Call the handler
	result := handleImageUpload(rr, req, mockStorage)
	
	// Check results
	if !result {
		t.Errorf("handleImageUpload returned false, expected true")
	}
	
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	
	// Check response body contains the URL
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if url, ok := response["url"]; !ok || url != mockURL {
		t.Errorf("Response URL incorrect, got %v, want %v", url, mockURL)
	}
}

// Test FormFile error
func TestHandleImageUpload_FormFileError(t *testing.T) {
	mockStorage := &MockImageStorage{}
	
	// Create a form without an image field
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add a different field instead
	w.WriteField("notimage", "this is not an image")
	w.Close()
	
	// Create test request
	req := httptest.NewRequest("POST", "/api/image/testpage", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Call the handler
	result := handleImageUpload(rr, req, mockStorage)
	
	// Should still return true as the handler processed the request, but with an error
	if !result {
		t.Errorf("handleImageUpload returned false for form error, expected true")
	}
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for form error: got %v want %v", 
			rr.Code, http.StatusBadRequest)
	}
}

// Test storage error
func TestHandleImageUpload_StorageError(t *testing.T) {
	// Setup mock storage with error
	storageErr := fmt.Errorf("storage error")
	mockStorage := &MockImageStorage{err: storageErr}
	
	// Create a test multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add image file to form
	fw, err := w.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	imageData := []byte("fake image data")
	fw.Write(imageData)
	w.Close()
	
	// Create test request
	req := httptest.NewRequest("POST", "/api/image/testpage", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Call the handler
	result := handleImageUpload(rr, req, mockStorage)
	
	// Should still return true as the handler processed the request, but with an error
	if !result {
		t.Errorf("handleImageUpload returned false for storage error, expected true")
	}
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code for storage error: got %v want %v", 
			rr.Code, http.StatusInternalServerError)
	}
}

func TestHandleImageUpload_InvalidPath(t *testing.T) {
	mockStorage := &MockImageStorage{}
	
	// Create test request with invalid path
	req := httptest.NewRequest("POST", "/api/wrong/path", nil)
	rr := httptest.NewRecorder()
	
	// Call the handler
	result := handleImageUpload(rr, req, mockStorage)
	
	// Check results - should return false for invalid path
	if result {
		t.Errorf("handleImageUpload returned true for invalid path, expected false")
	}
}

func TestAPIHandler_ImageUpload(t *testing.T) {
	// Setup mock storage
	mockURL := "/wiki/raw/images/test/12345.png"
	mockStorage := &MockImageStorage{imageURL: mockURL}
	
	// Create a test multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add image file to form
	fw, err := w.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	imageData := []byte("fake image data")
	fw.Write(imageData)
	w.Close()
	
	// Create test request
	req := httptest.NewRequest("POST", "/api/image/testpage", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	
	// Record response
	rr := httptest.NewRecorder()
	
	// Create handler and call it
	handler := apiHandler(innerAPIHandler, mockStorage)
	handler(rr, req)
	
	// Check response status
	if rr.Code != http.StatusOK {
		t.Errorf("API handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	
	// Check response body contains the URL
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if url, ok := response["url"]; !ok || url != mockURL {
		t.Errorf("Response URL incorrect, got %v, want %v", url, mockURL)
	}
}