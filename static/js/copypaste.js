document.addEventListener('DOMContentLoaded', function() {
  const editor = document.getElementById('wikiedit');
  if (!editor) return;

  function columnWidth(rows, columnIndex) {
      return Math.max.apply(null, rows.map(function(row) {
          if (typeof row[columnIndex] === 'undefined') {
              return 0
          } else {
              return row[columnIndex].length
          }
      }))
  }

  function looksLikeTable(data) {
      if (data.indexOf("\t") != -1) {
          return true
      }
      return false
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
      dialog.style.width = '450px';
      dialog.style.maxWidth = '95vw';
      dialog.innerHTML = `
        <h3 style="margin-top:0">Image Options</h3>
        <div style="display: flex; margin-bottom: 15px;">
          <div id="original-preview" style="flex: 1; margin-right: 10px;">
            <div style="font-weight: bold; margin-bottom: 5px; text-align: center;">Original</div>
            <div style="height: 150px; width: 100%; border: 1px dashed #ccc; display: flex; justify-content: center; align-items: center; overflow: hidden;">
              <img id="original-img" style="max-width: 100%; max-height: 100%; object-fit: contain;" />
            </div>
            <div id="original-dimensions" style="text-align: center; font-size: 12px; margin-top: 5px;"></div>
          </div>
          <div id="resized-preview" style="flex: 1; margin-left: 10px;">
            <div style="font-weight: bold; margin-bottom: 5px; text-align: center;">Preview</div>
            <div style="height: 150px; width: 100%; border: 1px dashed #ccc; display: flex; justify-content: center; align-items: center; overflow: hidden; position: relative;">
              <div style="position: absolute; width: 100%; height: 100%; display: flex; justify-content: center; align-items: center;">
                <img id="resized-img" style="max-width: 100%; max-height: 100%; object-fit: contain;" />
              </div>
            </div>
            <div id="resized-dimensions" style="text-align: center; font-size: 12px; margin-top: 5px;"></div>
          </div>
        </div>
        
        <div style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px;">Resize method:</label>
          <div style="display: flex; margin-bottom: 10px;">
            <label style="margin-right: 15px;">
              <input type="radio" name="resize-method" value="percent" checked> Percentage
            </label>
            <label>
              <input type="radio" name="resize-method" value="pixels"> Pixels
            </label>
          </div>
        </div>
        
        <div id="percent-controls" style="margin-bottom: 15px;">
          <label style="display: block; margin-bottom: 5px;">Scale:</label>
          <div style="display: flex; align-items: center;">
            <input type="range" id="scale-slider" min="10" max="100" value="100" style="flex: 1;">
            <span id="scale-value" style="margin-left: 10px; min-width: 40px;">100%</span>
          </div>
          <div style="display: flex; justify-content: space-between; margin-top: 10px;">
            <button class="preset-btn" data-scale="25" style="flex: 1; margin: 0 2px; padding: 4px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">25%</button>
            <button class="preset-btn" data-scale="50" style="flex: 1; margin: 0 2px; padding: 4px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">50%</button>
            <button class="preset-btn" data-scale="75" style="flex: 1; margin: 0 2px; padding: 4px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">75%</button>
            <button class="preset-btn" data-scale="100" style="flex: 1; margin: 0 2px; padding: 4px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">100%</button>
          </div>
        </div>
        
        <div id="pixel-controls" style="margin-bottom: 15px; display: none;">
          <div style="display: flex; margin-bottom: 10px;">
            <div style="flex: 1; margin-right: 5px;">
              <label style="display: block; margin-bottom: 5px;">Width (px):</label>
              <input type="number" id="image-width" placeholder="Width" style="width: 100%; padding: 8px; border-radius: 4px; border: 1px solid #ccc;" />
            </div>
            <div style="flex: 1; margin-left: 5px;">
              <label style="display: block; margin-bottom: 5px;">Height (px):</label>
              <input type="number" id="image-height" placeholder="Height" style="width: 100%; padding: 8px; border-radius: 4px; border: 1px solid #ccc;" />
            </div>
          </div>
          <div style="display: flex; align-items: center; margin-top: 5px;">
            <input type="checkbox" id="maintain-aspect" checked>
            <label for="maintain-aspect" style="margin-left: 5px;">Maintain aspect ratio</label>
          </div>
        </div>
        
        <div style="margin-top: 20px; display: flex; justify-content: space-between;">
          <button id="cancel-upload" style="padding: 8px 16px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">Cancel</button>
          <button id="upload-original" style="padding: 8px 16px; border-radius: 4px; border: 1px solid #ccc; background-color: #f0f0f0; cursor: pointer;">Upload Original</button>
          <button id="upload-resized" style="padding: 8px 16px; border-radius: 4px; border: none; background-color: #4a90e2; color: white; cursor: pointer;">Upload Resized</button>
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
      const originalImg = document.getElementById('original-img');
      const resizedImg = document.getElementById('resized-img');
      originalImg.src = imageUrl;
      resizedImg.src = imageUrl;
      
      let originalWidth = 0;
      let originalHeight = 0;
      let currentScale = 100;
      
      // Set original dimensions and init controls
      originalImg.onload = function() {
        originalWidth = originalImg.naturalWidth;
        originalHeight = originalImg.naturalHeight;
        
        // Display original dimensions
        document.getElementById('original-dimensions').textContent = 
          `${originalWidth} × ${originalHeight}`;
        
        // Initial resized dimensions are the same as original
        document.getElementById('resized-dimensions').textContent = 
          `${originalWidth} × ${originalHeight}`;
        
        // Set initial values for pixel inputs
        const widthInput = document.getElementById('image-width');
        const heightInput = document.getElementById('image-height');
        widthInput.value = originalWidth;
        heightInput.value = originalHeight;
        widthInput.placeholder = `Width (max ${originalWidth})`;
        heightInput.placeholder = `Height (max ${originalHeight})`;
        
        // Initialize the preview
        updateResizedPreview();
      };
      
      // Function to update the preview based on current settings
      function updateResizedPreview() {
        const isPercentMode = document.querySelector('input[name="resize-method"]:checked').value === 'percent';
        let newWidth, newHeight;
        
        if (isPercentMode) {
          // Percentage mode
          newWidth = Math.round(originalWidth * currentScale / 100);
          newHeight = Math.round(originalHeight * currentScale / 100);
        } else {
          // Pixel mode
          const widthInput = document.getElementById('image-width');
          const heightInput = document.getElementById('image-height');
          const maintainAspect = document.getElementById('maintain-aspect').checked;
          
          if (maintainAspect) {
            // If width was changed last, calculate height based on width
            if (document.activeElement === widthInput) {
              newWidth = parseInt(widthInput.value) || originalWidth;
              newHeight = Math.round(newWidth * originalHeight / originalWidth);
              heightInput.value = newHeight;
            } else {
              // Otherwise calculate width based on height
              newHeight = parseInt(heightInput.value) || originalHeight;
              newWidth = Math.round(newHeight * originalWidth / originalHeight);
              widthInput.value = newWidth;
            }
          } else {
            // Independent dimensions
            newWidth = parseInt(widthInput.value) || originalWidth;
            newHeight = parseInt(heightInput.value) || originalHeight;
          }
        }
        
        // Update resized dimensions display
        document.getElementById('resized-dimensions').textContent = `${newWidth} × ${newHeight}`;
        
        // Update the preview image scale
        const resizedImg = document.getElementById('resized-img');
        
        // Calculate scale factors
        const scaleX = newWidth / originalWidth;
        const scaleY = newHeight / originalHeight;
        
        // Apply CSS scales for the preview
        resizedImg.style.maxWidth = 'none';
        resizedImg.style.maxHeight = 'none';
        resizedImg.style.width = newWidth + 'px';
        resizedImg.style.height = newHeight + 'px';
        
        // For small images, center them in the preview container
        const previewContainer = document.querySelector('#resized-preview > div');
        const containerWidth = previewContainer.clientWidth;
        const containerHeight = previewContainer.clientHeight;
        
        // If the resized image is smaller than the container,
        // we need to center it with margins
        if (newWidth < containerWidth || newHeight < containerHeight) {
          resizedImg.style.marginLeft = 'auto';
          resizedImg.style.marginRight = 'auto';
          resizedImg.style.display = 'block';
        }
      }
      
      // Setup event handlers for the resize method radio buttons
      const resizeMethodRadios = document.querySelectorAll('input[name="resize-method"]');
      resizeMethodRadios.forEach(radio => {
        radio.addEventListener('change', function() {
          const isPercent = this.value === 'percent';
          document.getElementById('percent-controls').style.display = isPercent ? 'block' : 'none';
          document.getElementById('pixel-controls').style.display = isPercent ? 'none' : 'block';
          updateResizedPreview();
        });
      });
      
      // Setup event handlers for percentage slider
      const scaleSlider = document.getElementById('scale-slider');
      const scaleValue = document.getElementById('scale-value');
      scaleSlider.addEventListener('input', function() {
        currentScale = parseInt(this.value);
        scaleValue.textContent = `${currentScale}%`;
        updateResizedPreview();
      });
      
      // Setup event handlers for preset buttons
      const presetButtons = document.querySelectorAll('.preset-btn');
      presetButtons.forEach(btn => {
        btn.addEventListener('click', function() {
          currentScale = parseInt(this.dataset.scale);
          scaleSlider.value = currentScale;
          scaleValue.textContent = `${currentScale}%`;
          updateResizedPreview();
        });
      });
      
      // Setup event handlers for pixel inputs
      const widthInput = document.getElementById('image-width');
      const heightInput = document.getElementById('image-height');
      
      widthInput.addEventListener('input', updateResizedPreview);
      heightInput.addEventListener('input', updateResizedPreview);
      
      // Setup event handler for maintain aspect checkbox
      document.getElementById('maintain-aspect').addEventListener('change', updateResizedPreview);
      
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
        let width, height;
        const isPercentMode = document.querySelector('input[name="resize-method"]:checked').value === 'percent';
        
        if (isPercentMode) {
          // Calculate dimensions based on percentage
          width = Math.round(originalWidth * currentScale / 100);
          height = Math.round(originalHeight * currentScale / 100);
        } else {
          // Get dimensions from inputs
          width = parseInt(document.getElementById('image-width').value) || 0;
          height = parseInt(document.getElementById('image-height').value) || 0;
          
          if (width <= 0 && height <= 0) {
            alert('Please specify at least one valid dimension');
            return;
          }
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
      .then(response => {
        if (!response.ok) {
          throw new Error(`Server error: ${response.status} ${response.statusText}`);
        }
        return response.json();
      })
      .then(data => {
        if (data.url) {
          // Use standard markdown syntax for all images
          // The resized image URL already contains the proper dimensions from the server
          const markdownImage = `![Image](${data.url})`;
          
          // Insert at cursor position
          insertAtCursor(editor, markdownImage);
        } else {
          throw new Error('No image URL returned from server');
        }
        document.body.removeChild(statusEl);
      })
      .catch(error => {
        console.error('Error uploading image:', error);
        // Show error message to user
        statusEl.textContent = 'Error uploading image. Please try again.';
        statusEl.style.backgroundColor = '#ffeeee';
        statusEl.style.border = '1px solid #ff0000';
        // Auto-remove the error message after 5 seconds
        setTimeout(() => {
          if (document.body.contains(statusEl)) {
            document.body.removeChild(statusEl);
          }
        }, 5000);
      });
  }

  editor.addEventListener("paste", function(event) {
      // Handle clipboard data
      const items = (event.clipboardData || event.originalEvent.clipboardData).items;
      
      // Check for image data
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.indexOf('image') !== -1) {
          // We found an image!
          event.preventDefault();
          
          const blob = items[i].getAsFile();
          uploadImage(blob);
          return;
        }
      }
      
      // Handle table paste (existing functionality)
      var data = event.clipboardData.getData('text/plain').trim();

      if (looksLikeTable(data)) {
          try {
              var rows = data.split((/[\n\u0085\u2028\u2029]|\r\n?/g)).map(function(row) {
                  return row.split("\t")
              })
              var columnWidths = rows[0].map(function(column, columnIndex) {
                  return columnWidth(rows, columnIndex)
              })
              var markdownRows = rows.map(function(row, rowIndex) {
                  // | Name         | Title | Email Address  |
                  // |--------------|-------|----------------|
                  // | Jane Atler   | CEO   | jane@acme.com  |
                  // | John Doherty | CTO   | john@acme.com  |
                  // | Sally Smith  | CFO   | sally@acme.com |
                  return "| " + row.map(function(column, index) {
                      return column + Array(columnWidths[index] - column.length + 1).join(" ")
                  }).join(" | ") + " |"

              })
              markdownRows.splice(1, 0, "|" + columnWidths.map(function(width, index) {
                  return Array(columnWidths[index] + 3).join("-")
              }).join("|") + "|")

              // https://www.w3.org/TR/clipboard-apis/#the-paste-action
              // When pasting, the drag data store mode flag is read-only, hence calling
              // setData() from a paste event handler will not modify the data that is
              // inserted, and not modify the data on the clipboard.

              event.target.value += markdownRows.join("\n")

              event.preventDefault()
              return false

          } catch (e) {
              // Log the error out as it might be useful but assuming we've not called preventDefault 
              // the default action should just kick in
              console.log(e);
          }
      }
      return;
  });
});