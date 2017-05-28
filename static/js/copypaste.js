var editor = document.getElementById("wikiedit")

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

editor.addEventListener("paste", function(event) {
    var clipboard = event.clipboardData
    var data = clipboard.getData('text/plain').trim()

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
    return

})