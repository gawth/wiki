
$(document).ready(function() {

    $(".tag-label").click(function(event) {
        var jqxhr = $.getJSON("/api?tag=" + event.target.textContent, function() {
            })
            .done(function(data) {
            })
            .fail(function() {
            })
            .always(function() {
            });
    });

    var checkboxValues = JSON.parse(localStorage.getItem('checkboxValues')) || {};
    var $checkboxes = $("#menu :checkbox");

    $checkboxes.on("change", function(){
      $checkboxes.each(function(){
          checkboxValues[this.id] = this.checked;
        });
      localStorage.setItem("checkboxValues", JSON.stringify(checkboxValues));
    });

    $.each(checkboxValues, function(key, value) {
      $("#" + key).prop('checked', value);
    });
  });


