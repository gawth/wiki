
$(document).ready(function() {

    $(".tag-label").click(function(event) {
        console.log("Event " + event.target.textContent);
        var jqxhr = $.getJSON("/api?tag=" + event.target.textContent, function() {
                console.log("success");
            })
            .done(function(data) {
                console.log("second success:" + data);
            })
            .fail(function() {
                console.log("error");
            })
            .always(function() {
                console.log("finished");
            });
    });
});


