(function ($) {
  "use strict";

  function adjustPasteInput() {
      var h = $(window).height();
      if (h < 60) {
          h = 60;
      }
      $("#pasteinput").height(h - 10);
  }

  function syncChanges() {
    console.log("sync");
  }




  $(function () {
    adjustPasteInput();
    $(window).on("resize", adjustPasteInput);

    $("#pasteinput").on("input", syncChanges);
  });

}(jQuery));
