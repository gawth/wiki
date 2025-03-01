package main

import (
	"bytes"
	"image"
	"image/png"
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

func TestFileStorage_StoreResizedImage(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	origWikiDir := wikiDir
	wikiDir = tempDir + "/"
	defer func() { wikiDir = origWikiDir }()
	
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with a solid color for simplicity
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Each pixel is direct RGBA color
			offset := img.PixOffset(x, y)
			img.Pix[offset+0] = 255  // R
			img.Pix[offset+1] = 0    // G
			img.Pix[offset+2] = 0    // B
			img.Pix[offset+3] = 255  // A (fully opaque)
		}
	}
	
	// Encode the image
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
	imageData := buf.Bytes()
	
	// Test data
	testWiki := "testpage"
	testExt := ".png"
	
	// Create storage
	fs := fileStorage{}
	
	// Test resize to 50x50
	imageURL, err := fs.storeResizedImage(testWiki, imageData, testExt, 50, 50)
	if err != nil {
		t.Fatalf("Failed to store resized image: %v", err)
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
	
	// Verify the image dimensions
	savedFile := filepath.Join(imagesDir, files[0].Name())
	file, err := os.Open(savedFile)
	if err != nil {
		t.Fatalf("Failed to open saved image: %v", err)
	}
	defer file.Close()
	
	resizedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode saved image: %v", err)
	}
	
	bounds := resizedImg.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Image not resized correctly. Expected 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	
	// Test resize with only width (maintaining aspect ratio)
	imageURL, err = fs.storeResizedImage(testWiki, imageData, testExt, 30, 0)
	if err != nil {
		t.Fatalf("Failed to store width-only resized image: %v", err)
	}
	
	// Extract filename from URL
	urlParts := strings.Split(imageURL, "/")
	filename := urlParts[len(urlParts)-1]
	
	// Verify the proportional resize
	savedFile = filepath.Join(imagesDir, filename)
	file, err = os.Open(savedFile)
	if err != nil {
		t.Fatalf("Failed to open proportionally resized image: %v", err)
	}
	defer file.Close()
	
	resizedImg, _, err = image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode proportionally resized image: %v", err)
	}
	
	bounds = resizedImg.Bounds()
	if bounds.Dx() != 30 {
		t.Errorf("Width not resized correctly. Expected 30, got %d", bounds.Dx())
	}
	if bounds.Dy() != 30 {
		t.Errorf("Height not proportionally resized. Expected 30, got %d", bounds.Dy())
	}
}

func TestConfigurableStorage_StoreResizedImage(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	
	// Create test config
	config := StorageConfig{
		WikiDir: tempDir + "/",
	}
	
	// Create storage
	cs := NewConfigurableStorage(config)
	
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with a solid color for simplicity
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Each pixel is direct RGBA color
			offset := img.PixOffset(x, y)
			img.Pix[offset+0] = 0    // R
			img.Pix[offset+1] = 0    // G
			img.Pix[offset+2] = 255  // B
			img.Pix[offset+3] = 255  // A (fully opaque)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
	imageData := buf.Bytes()
	
	// Test data
	testWiki := "testpage"
	testExt := ".png"
	
	// Test storing a resized image
	imageURL, err := cs.storeResizedImage(testWiki, imageData, testExt, 40, 40)
	if err != nil {
		t.Fatalf("Failed to store resized image: %v", err)
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
	
	// Verify the image dimensions
	savedFile := filepath.Join(imagesDir, files[0].Name())
	file, err := os.Open(savedFile)
	if err != nil {
		t.Fatalf("Failed to open saved image: %v", err)
	}
	defer file.Close()
	
	resizedImg, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode saved image: %v", err)
	}
	
	bounds := resizedImg.Bounds()
	if bounds.Dx() != 40 || bounds.Dy() != 40 {
		t.Errorf("Image not resized correctly. Expected 40x40, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}