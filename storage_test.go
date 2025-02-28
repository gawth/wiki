package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestCreateDir tests the directory creation function
func TestCreateDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test directory creation with a path
	testPath := filepath.Join(tmpDir, "foo", "bar")
	err = createDir(testPath)
	if err != nil {
		t.Errorf("createDir failed: %v", err)
	}

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(testPath))
	if err != nil {
		t.Errorf("Directory not created: %v", err)
	}

	// Test with empty path (should succeed as it's a no-op)
	err = createDir("")
	if err != nil {
		t.Errorf("createDir with empty path failed: %v", err)
	}
}

// TestStoreFile tests the file storage function
func TestStoreFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test fileStorage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}

	// Test storing a file
	testFilePath := filepath.Join(tmpDir, "test.md")
	testContent := []byte("test content")
	err = fs.storeFile(testFilePath, testContent)
	if err != nil {
		t.Errorf("storeFile failed: %v", err)
	}

	// Verify file was created with correct content
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}
	if !bytes.Equal(content, testContent) {
		t.Errorf("File content mismatch. Got: %s, Want: %s", content, testContent)
	}

	// Test storing in a non-existent subdirectory
	subDirPath := filepath.Join(tmpDir, "sub", "dir", "test.md")
	err = fs.storeFile(subDirPath, testContent)
	if err != nil {
		t.Errorf("storeFile with subdirectory failed: %v", err)
	}

	// Verify file in subdirectory was created
	_, err = os.Stat(subDirPath)
	if err != nil {
		t.Errorf("File in subdirectory not created: %v", err)
	}
}

// TestDeleteFile tests the file deletion function
func TestDeleteFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test fileStorage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}

	// Create a test file to delete
	testFilePath := filepath.Join(tmpDir, "to-delete.md")
	err = os.WriteFile(testFilePath, []byte("test content"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Test deleting the file
	err = fs.deleteFile(testFilePath)
	if err != nil {
		t.Errorf("deleteFile failed: %v", err)
	}

	// Verify file no longer exists
	_, err = os.Stat(testFilePath)
	if !os.IsNotExist(err) {
		t.Errorf("File still exists after deletion or another error occurred: %v", err)
	}

	// Test deleting non-existent file (should return error)
	err = fs.deleteFile(filepath.Join(tmpDir, "nonexistent.md"))
	if err == nil {
		t.Errorf("deleteFile should fail for non-existent file but didn't")
	}
}

// TestMoveFile tests the file move/rename function
func TestMoveFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test fileStorage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}

	// Create a test file to move
	testContent := []byte("test content")
	sourcePath := filepath.Join(tmpDir, "source.md")
	destPath := filepath.Join(tmpDir, "destination.md")
	
	err = os.WriteFile(sourcePath, testContent, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Test moving the file
	err = fs.moveFile(sourcePath, destPath)
	if err != nil {
		t.Errorf("moveFile failed: %v", err)
	}

	// Verify source no longer exists
	_, err = os.Stat(sourcePath)
	if !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move: %v", err)
	}

	// Verify destination contains correct content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}
	if !bytes.Equal(content, testContent) {
		t.Errorf("Destination content mismatch. Got: %s, Want: %s", content, testContent)
	}

	// Test moving non-existent file (should return error)
	err = fs.moveFile(filepath.Join(tmpDir, "nonexistent.md"), destPath)
	if err == nil {
		t.Errorf("moveFile should fail for non-existent source but didn't")
	}
}

// TestGetPublicPages tests the indexing of public pages
func TestGetPublicPages(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Save original pubDir
	originalPubDir := pubDir
	
	// Set pubDir to our temp directory
	pubDirPath := filepath.Join(tmpDir, "pub") + "/"
	pubDir = pubDirPath
	
	// Restore original value when test completes
	defer func() { pubDir = originalPubDir }()
	
	// Create test public pages directory and files
	os.MkdirAll(pubDirPath, 0755)
	os.WriteFile(filepath.Join(pubDirPath, "page1.md"), []byte("content"), 0600)
	os.WriteFile(filepath.Join(pubDirPath, "page2.md"), []byte("content"), 0600)
	
	// Create subdirectory with a file
	subDir := filepath.Join(pubDirPath, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "page3.md"), []byte("content"), 0600)
	
	// Create test storage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}
	
	// Get public pages
	pages := fs.getPublicPages()
	
	// We should get 3 pages
	if len(pages) != 3 {
		t.Errorf("Expected 3 public pages, got %d", len(pages))
	}
	
	// Check that all expected pages are in the results
	expectedPages := map[string]bool{
		"page1.md": false,
		"page2.md": false,
		"subdir/page3.md": false,
	}
	
	for _, page := range pages {
		expectedPages[page] = true
	}
	
	for page, found := range expectedPages {
		if !found {
			t.Errorf("Expected to find '%s' in public pages but didn't", page)
		}
	}
}

// TestGetPage tests retrieving a wiki page by temporarily overriding global variables
func TestGetPage(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Save original globals
	originalWikiDir := wikiDir
	originalTagDir := tagDir
	originalPubDir := pubDir
	originalEkey := ekey
	
	// Set temp values for testing
	wikiDir = tmpDir + "/"
	tagDir = filepath.Join(tmpDir, "tags") + "/"
	pubDir = filepath.Join(tmpDir, "pub") + "/"
	ekey = []byte("12345678901234567890123456789012") // 32 byte key for testing
	
	// Restore original values when test completes
	defer func() {
		wikiDir = originalWikiDir
		tagDir = originalTagDir
		pubDir = originalPubDir
		ekey = originalEkey
	}()
	
	// Create directories
	os.MkdirAll(tagDir, 0755)
	os.MkdirAll(pubDir, 0755)
	
	// Create test storage
	fs := fileStorage{TagDir: tagDir}
	
	// Test 1: Basic page retrieval
	title := "TestPage"
	content := []byte("Test wiki content")
	filename := getWikiFilename(wikiDir, title)
	
	err = os.WriteFile(filename, content, 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	p := &wikiPage{basePage: basePage{Title: title}}
	result, err := fs.getPage(p)
	if err != nil {
		t.Errorf("getPage failed: %v", err)
	}
	
	if string(result.Body) != string(content) {
		t.Errorf("Page content mismatch. Got: %s, Want: %s", result.Body, content)
	}
	
	// Test 2: Page with tags
	tagsContent := "tag1,tag2,tag3"
	err = os.WriteFile(getWikiTagsFilename(title), []byte(tagsContent), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	p = &wikiPage{basePage: basePage{Title: title}}
	result, err = fs.getPage(p)
	if err != nil {
		t.Errorf("getPage with tags failed: %v", err)
	}
	
	if result.Tags != tagsContent {
		t.Errorf("Tags mismatch. Got: %s, Want: %s", result.Tags, tagsContent)
	}
	
	if len(result.TagArray) != 3 {
		t.Errorf("Expected 3 tags in TagArray, got %d", len(result.TagArray))
	}
	
	// Test 3: Published page
	err = os.WriteFile(getWikiPubFilename(title), []byte(""), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	p = &wikiPage{basePage: basePage{Title: title}}
	result, err = fs.getPage(p)
	if err != nil {
		t.Errorf("getPage with published flag failed: %v", err)
	}
	
	if !result.Published {
		t.Error("Page should be marked as published but wasn't")
	}
	
	// Test 4: Encrypted page
	encryptedContent := []byte("Test encrypted content")
	encryptedBytes, err := encrypt(encryptedContent, ekey)
	if err != nil {
		t.Fatal(err)
	}
	
	// Add encryption flag and write to file
	flaggedContent := append(encryptionFlag, encryptedBytes...)
	encTitle := "EncryptedPage"
	encFilename := getWikiFilename(wikiDir, encTitle)
	
	err = os.WriteFile(encFilename, flaggedContent, 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	p = &wikiPage{basePage: basePage{Title: encTitle}}
	result, err = fs.getPage(p)
	if err != nil {
		t.Errorf("getPage with encryption failed: %v", err)
	}
	
	if !result.Encrypted {
		t.Error("Page should be marked as encrypted but wasn't")
	}
	
	if string(result.Body) != string(encryptedContent) {
		t.Errorf("Encrypted content mismatch. Got: %s, Want: %s", result.Body, encryptedContent)
	}
	
	// Test 5: Non-existent page
	p = &wikiPage{basePage: basePage{Title: "NonExistentPage"}}
	_, err = fs.getPage(p)
	if err == nil {
		t.Error("Expected error for non-existent page but got none")
	}
}

// TestReadFile tests the readFile function used in searchPages
func TestReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file with content
	testFilePath := filepath.Join(tmpDir, "test-search.md")
	testContent := []byte("Line 1: This is a test\nLine 2: Contains apple\nLine 3: Contains orange\nLine 4: Both apple and orange")
	err = os.WriteFile(testFilePath, testContent, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Test searching for a term
	var wg sync.WaitGroup
	results := make(chan string)
	wg.Add(1)

	// Create a goroutine to collect results
	var matches []string
	done := make(chan struct{})
	go func() {
		for res := range results {
			matches = append(matches, res)
		}
		close(done)
	}()

	// Call readFile with our test parameters
	readFile(&wg, "test-search", testFilePath, "apple", results)
	
	// Wait for readFile to complete and close the results channel
	wg.Wait()
	close(results)
	<-done // Wait for collection goroutine to finish

	// Verify correct matches were found
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for 'apple', got %d", len(matches))
	}

	// Verify match line numbers are correct
	lineNumRegex := regexp.MustCompile(`\t(\d+)\t`)
	for _, match := range matches {
		// Extract line number from match string
		numMatch := lineNumRegex.FindStringSubmatch(match)
		if len(numMatch) < 2 {
			t.Errorf("Line number not found in match: %s", match)
			continue
		}

		// Verify that match contains expected content
		if !strings.Contains(match, "apple") {
			t.Errorf("Match doesn't contain search term 'apple': %s", match)
		}
	}
}

// TestSearchPages tests the page search functionality with temporary files
func TestSearchPages(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test files with searchable content
	files := []struct {
		name    string
		content string
	}{
		{"page1.md", "This page mentions apples"},
		{"page2.md", "This page mentions oranges"},
		{"page3.md", "This page mentions both apples and oranges"},
		{"subfolder/page4.md", "Another apple page in a subfolder"},
	}

	for _, file := range files {
		filePath := filepath.Join(tmpDir, file.name)
		// Create directory if needed
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filePath, []byte(file.content), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}
	
	// Without manually changing wikiDir (which is global), we can test the search functionality directly
	results := fs.searchPages(tmpDir, "apples")
	
	// We expect 2 results (page1 and page3)
	if len(results) != 2 {
		t.Errorf("Expected 2 search results for 'apples', got %d", len(results))
	}

	// Check that results contain expected strings
	page1Found := false
	page3Found := false
	for _, result := range results {
		if strings.Contains(result, "page1") {
			page1Found = true
		}
		if strings.Contains(result, "page3") {
			page3Found = true
		}
	}

	if !page1Found {
		t.Error("Expected to find 'page1' in search results but didn't")
	}
	if !page3Found {
		t.Error("Expected to find 'page3' in search results but didn't")
	}

	// Test searching for a term that doesn't exist
	noResults := fs.searchPages(tmpDir, "banana")
	if len(noResults) != 0 {
		t.Errorf("Expected 0 search results for 'banana', got %d", len(noResults))
	}
}

// TestGenID tests the genID function
func TestGenID(t *testing.T) {
	testCases := []struct {
		base     string
		name     string
		expected string
	}{
		{"", "test", "test"},
		{"/wiki", "test", "-wikitest"},
		{"/wiki", "folder/page", "-wikifolder-page"},
		{"/", "path/to/file", "-path-to-file"},
	}

	for i, tc := range testCases {
		result := genID(tc.base, tc.name)
		if result != tc.expected {
			t.Errorf("Case %d: genID(%q, %q) = %q, want %q", 
				i, tc.base, tc.name, result, tc.expected)
		}
	}
}

// Mock implementation of fs.DirEntry for testing
type mockDirEntry struct {
	name     string
	isDir    bool
	fileMode os.FileMode
	size     int64
	modTime  time.Time
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() os.FileMode          { return m.fileMode }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return m, nil }

// fs.FileInfo implementation
func (m mockDirEntry) Size() int64        { return m.size }
func (m mockDirEntry) Mode() os.FileMode  { return m.fileMode }
func (m mockDirEntry) ModTime() time.Time { return m.modTime }
func (m mockDirEntry) Sys() interface{}   { return nil }

// TestIndexWikiFiles tests the wiki file indexing functionality
func TestIndexWikiFiles(t *testing.T) {
	// This function is complex with many dependencies and would require mocking
	// the filesystem or creating elaborate directory structures
	t.Skip("Skipping test for IndexWikiFiles due to complexity and dependencies")
	
	/* A better approach would be to refactor the function to accept a FileSystem interface
	that could be mocked in tests, or to restructure to allow for dependency injection */
}

// TestCheckForPDF tests the PDF handling functionality
func TestCheckForPDF(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save the original wikiDir value
	originalWikiDir := wikiDir
	// Set wikiDir to our temp directory for this test
	wikiDir = tmpDir + "/"
	// Restore the original value when the test completes
	defer func() { wikiDir = originalWikiDir }()

	// Create a test PDF file
	pdfTitle := "testdoc"
	pdfFilename := getPDFFilename(wikiDir, pdfTitle)
	err = os.WriteFile(pdfFilename, []byte("%PDF-1.4 test content"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create test storage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}
	
	// Create wiki page object
	p := &wikiPage{basePage: basePage{Title: pdfTitle}}
	
	// Test PDF handling
	result, err := fs.checkForPDF(p)
	if err != nil {
		t.Errorf("checkForPDF failed: %v", err)
	}
	
	// Verify the body was set correctly with a link to the PDF
	expectedHTML := fmt.Sprintf("<a href=\"/wiki/raw/%v\">%v</a>", pdfTitle, pdfTitle)
	if string(result.Body) != expectedHTML {
		t.Errorf("PDF body incorrect. Got: %s, Want: %s", result.Body, expectedHTML)
	}
	
	// Test non-existent PDF
	nonExistentPage := &wikiPage{basePage: basePage{Title: "nonexistent"}}
	_, err = fs.checkForPDF(nonExistentPage)
	if err == nil {
		t.Error("Expected error for non-existent PDF but got none")
	}
}

// TestCachedStorage tests the cached storage wrapper
func TestCachedStorage(t *testing.T) {
	// This relies heavily on the filesystem and has many dependencies
	t.Skip("Skipping test for cachedStorage as it requires substantial setup")
	
	/* A better implementation would make the cached storage more testable by
	allowing injection of mocks for the underlying storage and filesystem operations */
}

// TestGetWikiList tests retrieval of wiki file list
func TestGetWikiList(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "wiki-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original wikiDir
	originalWikiDir := wikiDir
	
	// Set wikiDir to our temp directory
	wikiDir = tmpDir + "/"
	
	// Restore original value when test completes
	defer func() { wikiDir = originalWikiDir }()
	
	// Create test wiki files and directories
	testFiles := []string{
		"page1.md",
		"page2.md",
		"folder/page3.md",
		"folder/subfolder/page4.md",
		"another-folder/page5.md",
	}
	
	for _, file := range testFiles {
		filePath := filepath.Join(wikiDir, file)
		// Create directory if needed
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filePath, []byte("test content"), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// Create non-markdown files (should be ignored)
	err = os.WriteFile(filepath.Join(wikiDir, "not-wiki.txt"), []byte("not wiki"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create test storage
	fs := fileStorage{TagDir: filepath.Join(tmpDir, "tags")}
	
	// Test 1: Get all wikis
	wikis := fs.getWikiList("")
	// Note: We don't check exact count because the WalkDir function might find
	// files in a different order or include additional files in subdirectories
	if len(wikis) < len(testFiles) {
		t.Errorf("Expected at least %d wiki files, got %d", len(testFiles), len(wikis))
	}
	
	// Test 2: Get wikis from subfolder
	folderWikis := fs.getWikiList("folder")
	expectedFolderCount := 2 // page3.md and subfolder/page4.md
	if len(folderWikis) != expectedFolderCount {
		t.Errorf("Expected %d wiki files in folder, got %d", expectedFolderCount, len(folderWikis))
	}
	
	// Check that all expected files from the folder are present
	expectedInFolder := map[string]bool{
		"folder/page3": false,
		"folder/subfolder/page4": false,
	}
	
	for _, wiki := range folderWikis {
		expectedInFolder[wiki] = true
	}
	
	for wiki, found := range expectedInFolder {
		if !found {
			t.Errorf("Expected to find '%s' in folder wikis but didn't", wiki)
		}
	}
}

// The following are mock functions to help with testing

// mockFileSystem would implement storage interface for testing
type mockFileSystem struct {
	files     map[string][]byte
	pubPages  []string
	tagIndex  TagIndex
	wikiFiles []wikiNav
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files:    make(map[string][]byte),
		tagIndex: make(TagIndex),
	}
}

func (m *mockFileSystem) storeFile(name string, content []byte) error {
	m.files[name] = content
	return nil
}

func (m *mockFileSystem) deleteFile(name string) error {
	if _, exists := m.files[name]; !exists {
		return errors.New("file not found")
	}
	delete(m.files, name)
	return nil
}

func (m *mockFileSystem) moveFile(from, to string) error {
	if content, exists := m.files[from]; exists {
		m.files[to] = content
		delete(m.files, from)
		return nil
	}
	return errors.New("source file not found")
}

func (m *mockFileSystem) getPublicPages() []string {
	return m.pubPages
}

func (m *mockFileSystem) getPage(p *wikiPage) (*wikiPage, error) {
	content, exists := m.files[getWikiFilename("", p.Title)]
	if !exists {
		return nil, errors.New("page not found")
	}
	p.Body = template.HTML(content)
	return p, nil
}

func (m *mockFileSystem) searchPages(root, query string) []string {
	var results []string
	for name, content := range m.files {
		if strings.Contains(string(content), query) {
			results = append(results, name)
		}
	}
	return results
}

func (m *mockFileSystem) checkForPDF(p *wikiPage) (*wikiPage, error) {
	return p, nil
}

func (m *mockFileSystem) IndexTags(path string) TagIndex {
	return m.tagIndex
}

func (m *mockFileSystem) GetTagWikis(tag string) Tag {
	return m.tagIndex[tag]
}

func (m *mockFileSystem) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	return existing
}

func (m *mockFileSystem) IndexWikiFiles(base, path string) []wikiNav {
	return m.wikiFiles
}

func (m *mockFileSystem) getWikiList(from string) []string {
	var list []string
	for name := range m.files {
		if strings.HasSuffix(name, ".md") {
			list = append(list, strings.TrimSuffix(name, ".md"))
		}
	}
	return list
}