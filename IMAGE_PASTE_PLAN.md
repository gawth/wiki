# Wiki Image Features Implementation Plan

## Overview

This plan covers two key image-related features for the wiki editor:

1. **Image Paste** - Add clipboard image paste functionality to upload and insert images at cursor position
2. **Image Resize** - Enhance the image paste feature with resizing capabilities

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

## Image Resize Enhancement

### 1. Storage Interface Updates
- Extend the `storage` interface with a method for storing resized images
- Add a new method `storeResizedImage` to handle image resizing and saving

```go
// Add to storage.go - update interface
type storage interface {
    // Existing methods...
    storeImage(wikiTitle string, imageData []byte, extension string) (string, error)
    storeResizedImage(wikiTitle string, imageData []byte, extension string, width, height int) (string, error)
}

// Implementation for fileStorage
func (fst *fileStorage) storeResizedImage(wikiTitle string, imageData []byte, extension string, width, height int) (string, error) {
    // Create images directory if needed
    imagesDir := filepath.Join(wikiDir, "images", wikiTitle)
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return "", err
    }
    
    // Decode image data
    reader := bytes.NewReader(imageData)
    var img image.Image
    var err error
    
    switch strings.ToLower(extension) {
    case ".jpg", ".jpeg":
        img, err = jpeg.Decode(reader)
    case ".png":
        img, err = png.Decode(reader)
    default:
        // For other formats, use the generic image decoder
        img, _, err = image.Decode(reader)
    }
    
    if err != nil {
        return "", fmt.Errorf("failed to decode image: %v", err)
    }
    
    // Resize the image while maintaining aspect ratio
    var resized *image.NRGBA
    if width > 0 && height > 0 {
        resized = imaging.Resize(img, width, height, imaging.Lanczos)
    } else if width > 0 {
        resized = imaging.Resize(img, width, 0, imaging.Lanczos)
    } else if height > 0 {
        resized = imaging.Resize(img, 0, height, imaging.Lanczos)
    } else {
        return "", fmt.Errorf("at least one dimension (width or height) must be specified")
    }
    
    // Generate unique filename with timestamp and dimensions
    timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
    filename := fmt.Sprintf("%s_%dx%d%s", timestamp, resized.Bounds().Dx(), resized.Bounds().Dy(), extension)
    filepath := filepath.Join(imagesDir, filename)
    
    // Create the output file
    outFile, err := os.Create(filepath)
    if err != nil {
        return "", fmt.Errorf("failed to create output file: %v", err)
    }
    defer outFile.Close()
    
    // Save the resized image
    switch strings.ToLower(extension) {
    case ".jpg", ".jpeg":
        err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 90})
    case ".png":
        err = png.Encode(outFile, resized)
    default:
        // Default to PNG for unknown formats
        err = png.Encode(outFile, resized)
    }
    
    if err != nil {
        return "", fmt.Errorf("failed to encode resized image: %v", err)
    }
    
    // Return URL to client
    imageURL := fmt.Sprintf("/wiki/raw/images/%s/%s", wikiTitle, filename)
    return imageURL, nil
}
```

### 2. Update API Endpoint for Image Resize Support
- Modify the existing image upload API to accept width and height parameters
- Process these parameters and call the appropriate storage method

```go
// Update in api.go
func handleImageUpload(w http.ResponseWriter, r *http.Request, s storage) bool {
    // Existing code...
    
    var imageURL string
    
    // Check if resize parameters were provided
    widthStr := r.FormValue("width")
    heightStr := r.FormValue("height")
    
    if widthStr != "" || heightStr != "" {
        // Parse resize dimensions
        width, height := 0, 0
        if widthStr != "" {
            if w, err := strconv.Atoi(widthStr); err == nil {
                width = w
            }
        }
        if heightStr != "" {
            if h, err := strconv.Atoi(heightStr); err == nil {
                height = h
            }
        }
        
        // Store resized image
        imageURL, err = s.storeResizedImage(wikiTitle, imageData, fileExt, width, height)
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
```

### 3. Enhanced JavaScript for Image Resizing UI
- Update `static/js/copypaste.js` to add a resize dialog
- Provide width and height inputs with original dimension display
- Add preview functionality for the resized image

```javascript
// Update in static/js/copypaste.js
function uploadImage(blob) {
    // Get wiki title from URL
    const pathParts = window.location.pathname.split('/');
    const editIndex = pathParts.indexOf('edit');
    const wikiTitle = editIndex >= 0 && editIndex < pathParts.length - 1 ? pathParts[editIndex + 1] : '';
    
    if (!wikiTitle) {
        console.error('Could not determine wiki title from URL');
        return;
    }
    
    // Create image resize dialog
    const dialog = document.createElement('div');
    dialog.style.position = 'fixed';
    dialog.style.top = '50%';
    dialog.style.left = '50%';
    dialog.style.transform = 'translate(-50%, -50%)';
    dialog.style.backgroundColor = '#fff';
    dialog.style.padding = '20px';
    dialog.style.borderRadius = '8px';
    dialog.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
    dialog.style.zIndex = '1000';
    dialog.style.width = '400px';
    dialog.style.maxWidth = '95vw';
    dialog.innerHTML = `
        <h3 style="margin-top:0">Image Options</h3>
        <div id="image-preview" style="max-width: 100%; height: 150px; margin-bottom: 10px; display: flex; justify-content: center; align-items: center; border: 1px dashed #ccc;">
            <img style="max-width: 100%; max-height: 100%;" />
        </div>
        <div style="margin-bottom: 15px;">
            <label style="display: block; margin-bottom: 5px;">Width (px):</label>
            <input type="number" id="image-width" placeholder="Original width" style="width: 100%; padding: 8px; border-radius: 4px; border: 1px solid #ccc;" />
        </div>
        <div style="margin-bottom: 15px;">
            <label style="display: block; margin-bottom: 5px;">Height (px):</label>
            <input type="number" id="image-height" placeholder="Original height" style="width: 100%; padding: 8px; border-radius: 4px; border: 1px solid #ccc;" />
        </div>
        <div style="margin-top: 20px; display: flex; justify-content: space-between;">
            <button id="cancel-upload" style="padding: 8px 16px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">Cancel</button>
            <button id="upload-original" style="padding: 8px 16px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">Upload Original</button>
            <button id="upload-resized" style="padding: 8px 16px; border-radius: 4px; border: none; background-color: #4a90e2; color: white; cursor: pointer;">Upload with Resize</button>
        </div>
    `;
    
    // Add overlay
    const overlay = document.createElement('div');
    overlay.style.position = 'fixed';
    overlay.style.top = '0';
    overlay.style.left = '0';
    overlay.style.width = '100%';
    overlay.style.height = '100%';
    overlay.style.backgroundColor = 'rgba(0,0,0,0.5)';
    overlay.style.zIndex = '999';
    
    document.body.appendChild(overlay);
    document.body.appendChild(dialog);
    
    // Load image preview
    const imageUrl = URL.createObjectURL(blob);
    const previewImg = dialog.querySelector('#image-preview img');
    previewImg.src = imageUrl;
    
    // Set original dimensions
    previewImg.onload = function() {
        const widthInput = document.getElementById('image-width');
        const heightInput = document.getElementById('image-height');
        widthInput.placeholder = `Original width (${previewImg.naturalWidth}px)`;
        heightInput.placeholder = `Original height (${previewImg.naturalHeight}px)`;
    };
    
    // Handle cancel
    document.getElementById('cancel-upload').addEventListener('click', function() {
        document.body.removeChild(dialog);
        document.body.removeChild(overlay);
        URL.revokeObjectURL(imageUrl);
    });
    
    // Handle upload original
    document.getElementById('upload-original').addEventListener('click', function() {
        performUpload(blob, wikiTitle);
        document.body.removeChild(dialog);
        document.body.removeChild(overlay);
        URL.revokeObjectURL(imageUrl);
    });
    
    // Handle upload resized
    document.getElementById('upload-resized').addEventListener('click', function() {
        const width = document.getElementById('image-width').value;
        const height = document.getElementById('image-height').value;
        
        if (!width && !height) {
            alert('Please specify at least one dimension (width or height)');
            return;
        }
        
        performUpload(blob, wikiTitle, width, height);
        document.body.removeChild(dialog);
        document.body.removeChild(overlay);
        URL.revokeObjectURL(imageUrl);
    });
}

function performUpload(blob, wikiTitle, width, height) {
    // Show upload status
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
    
    // Create FormData and append the image
    const formData = new FormData();
    formData.append('image', blob, 'clipboard-image.png');
    
    // Add resize parameters if provided
    if (width) formData.append('width', width);
    if (height) formData.append('height', height);
    
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
```

### 4. Tests for Image Resizing

Add tests for the resizing functionality in `storage_image_test.go`:

```go
func TestFileStorage_StoreResizedImage(t *testing.T) {
    // Create a temp directory for testing
    tempDir := t.TempDir()
    origWikiDir := wikiDir
    wikiDir = tempDir + "/"
    defer func() { wikiDir = origWikiDir }()
    
    // Create a test image
    img := image.NewRGBA(image.Rect(0, 0, 100, 100))
    // Fill with a solid color
    for y := 0; y < 100; y++ {
        for x := 0; x < 100; x++ {
            offset := img.PixOffset(x, y)
            img.Pix[offset+0] = 255  // R
            img.Pix[offset+1] = 0    // G
            img.Pix[offset+2] = 0    // B
            img.Pix[offset+3] = 255  // A
        }
    }
    
    // Encode the image
    var buf bytes.Buffer
    if err := png.Encode(&buf, img); err != nil {
        t.Fatalf("Failed to encode test image: %v", err)
    }
    imageData := buf.Bytes()
    
    // Create storage
    fs := fileStorage{}
    
    // Test resize to 50x50
    imageURL, err := fs.storeResizedImage("testpage", imageData, ".png", 50, 50)
    if err != nil {
        t.Fatalf("Failed to store resized image: %v", err)
    }
    
    // Verify the image is created and has correct dimensions
    // (rest of the test implementation)
}
```

## Future Enhancements
- Add progress indicator for large image uploads
- Add image cropping functionality
- Add drag-and-drop image support with resize options
- Add direct file selection for uploads
- Implement image rotation options
- Add thumbnail generation with links to full-size images