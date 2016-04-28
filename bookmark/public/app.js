// jshint asi:true
// jshint unused:true
(function() {
  "use strict"

  var bookmarklet = 'javascript:void%20function(){var%20e=new%20XMLHttpRequest;e.open(%22POST%22,%22' + location.origin +'/api/bookmarks%22,!0),e.setRequestHeader(%22Content-Type%22,%22application/json%22),e.onerror=function(){alert(%22cannot%20make%20CORS%20request%22)},e.onload=function(){console.log(%22completed%22,e)},e.send(JSON.stringify({url:location.href}))}();'


  var Bookmarks = {
    list: function() {
      return m.request({
        method: "GET",
        url: "/api/bookmarks",
      }).then(function(resp) {
        return resp.bookmarks;
      })
    },
  }


  var listing = {
    controller: function() {
      this.bookmarks = Bookmarks.list()
    },
    view: function(ctrl) {
      return m("div", [
        m("a", {href: bookmarklet}, "bookmarklet"),
        ctrl.bookmarks().length + " bookmarks",
        ctrl.bookmarks().map(function(b) {
          return m("div", [
              m("a", {href: b.url}, b.title),
          ])
        })
      ])
    },
  }

  m.route.mode = 'pathname'

  window.onload = function() {
    m.route(document.getElementById("application"), "/", {
      '/': listing,
    })
  }

}())
