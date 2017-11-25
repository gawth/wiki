console.log("Hello Alan");
console.log($.support);

$(document).ready(function() {
    console.log('ready!');

    $(".tag").click(function() {
        var jqxhr = $.get("/api", function() {
                alert("success");
            })
            .done(function() {
                alert("second success");
            })
            .fail(function() {
                alert("error");
            })
            .always(function() {
                alert("finished");
            });
    });
});