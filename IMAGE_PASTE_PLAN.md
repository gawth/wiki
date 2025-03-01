# Image Paste Feature Implementation Plan

## Overview
Add clipboard image paste functionality to the wiki editor. When a user pastes an image from clipboard while editing a wiki page, the image will be uploaded to the server, stored in a wiki-specific directory, and a markdown image tag will be inserted at the cursor position.

## Implementation Steps

### 1. Storage Interface Updates
- Extend the `storage` interface with a method for storing images
- Add a new method `storeImage` to handle saving image data with proper filename generation
- Implement this method in both `fileStorage` and `s3Storage` (if used)

```go
// Add to storage.go - update interface
type storage interface {
    // Existing methods...
    storeImage(wikiTitle string, imageData []byte, extension string) (string, error)
}

// Implementation for fileStorage
func (s fileStorage) storeImage(wikiTitle string, imageData []byte, extension string) (string, error) {
    // Create images directory if needed
    imagesDir := filepath.Join(wikiDir, "images", wikiTitle)
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return "", err
    }
    
    // Generate unique filename with timestamp
    timestamp := time.Now().UnixNano()
    filename := fmt.Sprintf("%d%s", timestamp, extension)
    filepath := filepath.Join(imagesDir, filename)
    
    // Save file
    if err := os.WriteFile(filepath, imageData, 0644); err != nil {
        return "", err
    }
    
    // Return URL to client
    imageURL := fmt.Sprintf("/wiki/raw/images/%s/%s", wikiTitle, filename)
    return imageURL, nil
}
```

### 2. API Endpoint for Image Upload
- Add a new API endpoint in `api.go` for handling image uploads
- Process multipart form data to extract the image
- Use the storage interface to save the image
- Return a JSON response with the image URL

```go
// Add to api.go
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
    
    // Store image using storage interface
    imageURL, err := s.storeImage(wikiTitle, imageData, fileExt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return true
    }
    
    // Return URL to client
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"url": imageURL})
    
    return true
}

// Update innerAPIHandler to check for image uploads
func innerAPIHandler(w http.ResponseWriter, r *http.Request, s storage) {
    // Add this before other handlers
    if strings.HasPrefix(r.URL.Path, "/api/image/") {
        if handleImageUpload(w, r, s) {
            return
        }
    }
    
    // Existing code...
}
```

### 3. Update File Serving in wiki.go
- Ensure the `/wiki/raw/` handler can serve images from the images directory
- No changes needed if the existing file server is already set up correctly

### 4. Add JavaScript for Clipboard Image Handling
- Create or update `static/js/copypaste.js`
- Add event listener for paste events on the editor textarea
- Check for image data in the clipboard
- Upload image using fetch API
- Insert markdown at cursor position

```javascript
document.addEventListener('DOMContentLoaded', function() {
  const editor = document.getElementById('wikiedit');
  if (!editor) return;

  editor.addEventListener('paste', function(e) {
    // Handle clipboard data
    const items = (e.clipboardData || e.originalEvent.clipboardData).items;
    
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        // We found an image!
        e.preventDefault();
        
        const blob = items[i].getAsFile();
        uploadImage(blob);
        return;
      }
    }
  });

  function uploadImage(blob) {
    // Show upload status (optional)
    const statusEl = document.createElement('div');
    statusEl.textContent = 'Uploading image...';
    statusEl.style.position = 'fixed';
    statusEl.style.top = '10px';
    statusEl.style.right = '10px';
    statusEl.style.padding = '8px 16px';
    statusEl.style.backgroundColor = '#f0f0f0';
    statusEl.style.border = '1px solid #ccc';
    statusEl.style.borderRadius = '4px';
    document.body.appendChild(statusEl);
    
    // Get wiki title from URL or form action
    const wikiTitle = window.location.pathname.split('/edit/')[1];
    
    // Create FormData and append the image
    const formData = new FormData();
    formData.append('image', blob, 'clipboard-image.png');
    
    // Send to server
    fetch('/api/image/' + wikiTitle, {
      method: 'POST',
      body: formData
    })
    .then(response => response.json())
    .then(data => {
      if (data.url) {
        // Insert markdown for image at cursor position
        insertAtCursor(editor, `![Image](${data.url})`);
      }
      document.body.removeChild(statusEl);
    })
    .catch(error => {
      console.error('Error uploading image:', error);
      document.body.removeChild(statusEl);
    });
  }

  function insertAtCursor(textarea, text) {
    const startPos = textarea.selectionStart;
    const endPos = textarea.selectionEnd;
    
    textarea.value = 
      textarea.value.substring(0, startPos) + 
      text + 
      textarea.value.substring(endPos);
    
    // Set cursor position after inserted text
    textarea.selectionStart = textarea.selectionEnd = startPos + text.length;
    textarea.focus();
  }
});
```

### 5. Update edit.html
- Make sure the JavaScript file is included in the template

```html
<!-- Add if not already present in edit.html -->
<script src="/static/js/copypaste.js"></script>
```

## Testing

### 1. Unit Tests

Create `storage_image_test.go` to test the image storage functionality:

```go
package main

import (
    "bytes"
    "fmt"
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
```

### 2. API Endpoint Tests

Create `api_image_test.go` to test the image upload API endpoint:

```go
package main

import (
    "bytes"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "strings"
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
```

### 3. Manual Testing

1. Build and run the application: `go run wiki.go`
2. Open a wiki page in edit mode
3. Copy an image to clipboard (e.g., from a screenshot tool)
4. Click in the editor and paste (Ctrl+V or Cmd+V)
5. Verify the image uploads and markdown is inserted
6. Save the wiki page and verify the image displays correctly

## Future Enhancements
- Add progress indicator for large image uploads
- Allow for image resizing before upload
- Add drag-and-drop image support
- Add direct file selection for uploads