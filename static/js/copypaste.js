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
      
      // Get wiki title from URL
      const pathParts = window.location.pathname.split('/');
      const editIndex = pathParts.indexOf('edit');
      const wikiTitle = editIndex >= 0 && editIndex < pathParts.length - 1 ? pathParts[editIndex + 1] : '';
      
      if (!wikiTitle) {
          console.error('Could not determine wiki title from URL');
          document.body.removeChild(statusEl);
          return;
      }
      
      // Create FormData and append the image
      const formData = new FormData();
      formData.append('image', blob, 'clipboard-image.png');
      
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