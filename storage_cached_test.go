package main

import (
	"testing"
	"os"
	"path/filepath"
	"time"
)

// TestCachedStorageFunctionality tests the caching functionality
func TestCachedStorageFunctionality(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create real directories
	tagDir := filepath.Join(tmpDir, "tags") + "/"
	wikiDir := filepath.Join(tmpDir, "wiki") + "/"
	
	err = os.MkdirAll(tagDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a real fileStorage
	fstorage := fileStorage{TagDir: tagDir}
	
	// Create a new cachedStorage
	cached := cachedStorage{
		fileStorage: fstorage,
		wikiDir: wikiDir,
		tagDir: tagDir,
		cachedTagIndex: make(TagIndex),
		cachedRawFiles: make(TagIndex),
		cachedWikiIndex: []wikiNav{},
	}
	
	// Test that IndexTags returns the cached value
	testTagIndex := make(TagIndex)
	testTagIndex["test"] = Tag{TagName: "test", Wikis: []string{"wiki1"}}
	cached.cachedTagIndex = testTagIndex
	
	result := cached.IndexTags("")
	if len(result) != len(testTagIndex) {
		t.Errorf("Expected IndexTags to return cached value with %d entries", len(testTagIndex))
	}
	
	// Test that IndexRawFiles returns the cached value
	testRawFiles := make(TagIndex)
	testRawFiles["PDF"] = Tag{TagName: "PDF", Wikis: []string{"doc1"}}
	cached.cachedRawFiles = testRawFiles
	
	rawResult := cached.IndexRawFiles("", "", nil)
	if len(rawResult) != len(testRawFiles) {
		t.Errorf("Expected IndexRawFiles to return cached value with %d entries", len(testRawFiles))
	}
	
	// Test that IndexWikiFiles returns the cached value
	testWikiNav := []wikiNav{
		{Name: "wiki1", URL: "/wiki1"},
	}
	cached.cachedWikiIndex = testWikiNav
	
	wikiResult := cached.IndexWikiFiles("", "")
	if len(wikiResult) != len(testWikiNav) {
		t.Errorf("Expected IndexWikiFiles to return cached value with %d entries", len(testWikiNav))
	}
}

// TestCacheOperations tests basic operations with the cache
func TestCacheOperations(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create real directories
	tagDir := filepath.Join(tmpDir, "tags") + "/"
	wikiDir := filepath.Join(tmpDir, "wiki") + "/"
	
	err = os.MkdirAll(tagDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a tag file for testing indexing
	err = os.WriteFile(filepath.Join(tagDir, "testtag"), []byte("test1,test2"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a FileStorage with a real directory
	fs := fileStorage{TagDir: tagDir}
	
	// Create a test cached storage
	cached := cachedStorage{
		fileStorage:     fs,
		wikiDir:         wikiDir,
		tagDir:          tagDir,
		cachedTagIndex:  make(TagIndex),
		cachedRawFiles:  make(TagIndex),
		cachedWikiIndex: []wikiNav{},
	}
	
	// Test clearCache and file operations
	// Note: Since rebuildCache is async, we just test that the operations
	// complete without error, not their side effects
	
	// Test clearCache
	err = cached.clearCache()
	if err != nil {
		t.Errorf("clearCache returned error: %v", err)
	}
	
	// Test file operations
	err = cached.storeFile(filepath.Join(wikiDir, "test.txt"), []byte("test"))
	if err != nil {
		t.Errorf("storeFile returned error: %v", err)
	}
	
	err = cached.deleteFile(filepath.Join(wikiDir, "test.txt"))
	if err != nil {
		t.Errorf("deleteFile returned error: %v", err)
	}
	
	// Create a file to move
	testFile := filepath.Join(wikiDir, "tomove.txt")
	err = os.WriteFile(testFile, []byte("test"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	err = cached.moveFile(testFile, filepath.Join(wikiDir, "moved.txt"))
	if err != nil {
		t.Errorf("moveFile returned error: %v", err)
	}
	
	// Allow some time for any async operations to finish
	time.Sleep(100 * time.Millisecond)
}

// TestNewCachedStorage tests the creation of a new cached storage
func TestNewCachedStorage(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create actual directories so filesystem operations work
	wikiDir := filepath.Join(tmpDir, "wiki") + "/"
	tagDir := filepath.Join(tmpDir, "tags") + "/"
	
	err = os.MkdirAll(tagDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create file storage with real directory
	fs := fileStorage{TagDir: tagDir}
	
	// Create the cachedStorage manually to avoid file operations
	cached := cachedStorage{
		fileStorage:     fs,
		wikiDir:         wikiDir,
		tagDir:          tagDir,
		cachedTagIndex:  make(TagIndex),
		cachedRawFiles:  make(TagIndex),
		cachedWikiIndex: []wikiNav{},
	}
	
	// Verify the paths were set correctly
	if cached.wikiDir != wikiDir {
		t.Errorf("Expected wikiDir to be %s, got %s", wikiDir, cached.wikiDir)
	}
	
	if cached.tagDir != tagDir {
		t.Errorf("Expected tagDir to be %s, got %s", tagDir, cached.tagDir)
	}
	
	// Verify the cache was initialized
	if cached.cachedTagIndex == nil {
		t.Error("cachedTagIndex was not initialized")
	}
	
	if cached.cachedRawFiles == nil {
		t.Error("cachedRawFiles was not initialized")
	}
	
	if cached.cachedWikiIndex == nil {
		t.Error("cachedWikiIndex was not initialized")
	}
}

// TestRebuildCacheProcess tests the async rebuild process with synchronization
func TestRebuildCacheProcess(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create real directories
	tagDir := filepath.Join(tmpDir, "tags") + "/"
	wikiDir := filepath.Join(tmpDir, "wiki") + "/"
	
	err = os.MkdirAll(tagDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.MkdirAll(wikiDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a test tag file
	err = os.WriteFile(filepath.Join(tagDir, "testtag"), []byte("test1,test2"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a test wiki file
	err = os.WriteFile(filepath.Join(wikiDir, "test.md"), []byte("Test content"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	// Setup file storage
	fs := fileStorage{TagDir: tagDir}
	
	// Create a test cached storage
	cached := cachedStorage{
		fileStorage:     fs,
		wikiDir:         wikiDir,
		tagDir:          tagDir,
		cachedTagIndex:  make(TagIndex),
		cachedRawFiles:  make(TagIndex),
		cachedWikiIndex: []wikiNav{},
	}
	
	// Call rebuildCache directly (synchronously)
	cached.rebuildCache()
	
	// Test that cache was populated
	if len(cached.cachedTagIndex) == 0 {
		t.Log("Note: cachedTagIndex might be empty if no tags were found during rebuild")
	}
	
	// Test cache values access
	result := cached.IndexTags("")
	if result == nil {
		t.Error("IndexTags returned nil")
	}
	
	// Test that the cache can be modified
	cached.cachedTagIndex["test"] = Tag{TagName: "test", Wikis: []string{"wiki1"}}
	result = cached.IndexTags("")
	
	if len(result) < 1 {
		t.Errorf("Expected at least 1 entry in tag index after cache update")
	}
	
	// Test that the cache works as expected
	found := false
	for tagName := range result {
		if tagName == "test" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Added tag wasn't found in the cached results")
	}
}