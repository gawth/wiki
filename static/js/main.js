
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

    var todoValues = JSON.parse(localStorage.getItem('todoValues')) || {};
    var $todos = $("#todos :checkbox");

    $todos.on("change", function(){
      $todos.each(function(){
          todoValues[this.id] = this.checked;
        });
      localStorage.setItem("todoValues", JSON.stringify(todoValues));
    });

    $.each(todoValues, function(key, value) {
      $("#" + key).prop('checked', value);
    });
  });


