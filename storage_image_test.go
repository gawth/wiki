package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStorage_StoreImage(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	origWikiDir := wikiDir
	wikiDir = tempDir + "/"
	defer func() { wikiDir = origWikiDir }()
	
	// Test data
	testWiki := "testpage"
	testImage := []byte("fake image data")
	testExt := ".png"
	
	// Create storage
	fs := fileStorage{}
	
	// Test storing an image
	imageURL, err := fs.storeImage(testWiki, testImage, testExt)
	if err != nil {
		t.Fatalf("Failed to store image: %v", err)
	}
	
	// Verify URL format
	if !strings.HasPrefix(imageURL, "/wiki/raw/images/") {
		t.Errorf("Image URL has incorrect format: %s", imageURL)
	}
	
	// Verify file was created
	imagesDir := filepath.Join(tempDir, "images", testWiki)
	files, err := os.ReadDir(imagesDir)
	if err != nil {
		t.Fatalf("Failed to read images directory: %v", err)
	}
	
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}
	
	// Verify file content
	savedFile := filepath.Join(imagesDir, files[0].Name())
	content, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("Failed to read saved image: %v", err)
	}
	
	if !bytes.Equal(content, testImage) {
		t.Errorf("Saved image content doesn't match original")
	}
}

// Test file storage with directory creation error
func TestFileStorage_StoreImage_DirectoryError(t *testing.T) {
	// Save original and set to non-writable directory
	origWikiDir := wikiDir
	wikiDir = "/non/existent/directory/"
	defer func() { wikiDir = origWikiDir }()
	
	// Create storage
	fs := fileStorage{}
	
	// Test storing an image - should fail
	_, err := fs.storeImage("testpage", []byte("test"), ".png")
	if err == nil {
		t.Errorf("Expected directory creation error, but got no error")
	}
}

func TestConfigurableStorage_StoreImage(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	
	// Create test config
	config := StorageConfig{
		WikiDir: tempDir + "/",
	}
	
	// Create storage
	cs := NewConfigurableStorage(config)
	
	// Test data
	testWiki := "testpage"
	testImage := []byte("fake image data")
	testExt := ".png"
	
	// Test storing an image
	imageURL, err := cs.storeImage(testWiki, testImage, testExt)
	if err != nil {
		t.Fatalf("Failed to store image: %v", err)
	}
	
	// Verify URL format
	if !strings.HasPrefix(imageURL, "/wiki/raw/images/") {
		t.Errorf("Image URL has incorrect format: %s", imageURL)
	}
	
	// Verify file was created
	imagesDir := filepath.Join(tempDir, "images", testWiki)
	files, err := os.ReadDir(imagesDir)
	if err != nil {
		t.Fatalf("Failed to read images directory: %v", err)
	}
	
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}
	
	// Verify file content
	savedFile := filepath.Join(imagesDir, files[0].Name())
	content, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("Failed to read saved image: %v", err)
	}
	
	if !bytes.Equal(content, testImage) {
		t.Errorf("Saved image content doesn't match original")
	}
}