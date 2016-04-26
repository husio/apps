// jshint asi:true
// jshint unused:true
(function() {
  "use strict"

  var storage = window.localStorage

  var Entry = {
    list: function() {
      return  {
        entries: [
          {id: 213332, url: "http://blog.codinghorror.com/the-hugging-will-continue-until-morale-improves/", title: "The Hugging Will Continue Until Morale Improves"},
          {id: 421421, url: "https://blog.gopheracademy.com/advent-2015/ssh-server-in-go/", title: "Writing an SSH server in Go"},
          {id: 536262, url: "http://blog.codinghorror.com/welcome-to-the-internet-of-compromised-things/", title: "Welcome to The Internet of Compromised Things"},
          {id: 942174, url: "http://feedproxy.google.com/~r/HighScalability/~3/Yl0tVEk8fcc/the-joy-of-deploying-apache-storm-on-docker-swarm.html", title: "The Joy of Deploying Apache Storm on Docker Swarm"},
          {id: 928414, url: "http://feedproxy.google.com/~r/HighScalability/~3/IZLjlg8ua9g/stuff-the-internet-says-on-scalability-for-april-22nd-2016.html", title: "Stuff The Internet Says On Scalability For April 22nd, 2016"},
          {id: 991422, url: "http://golangweekly.com/issues/106", title: "Go Newsletter Issue #106"},
          {id: 848581, url: "http://feedproxy.google.com/~r/HighScalability/~3/IftH5Efwms4/how-twitter-handles-3000-images-per-second.html", title: "How Twitter Handles 3,000 Images Per Second"},
          {id: 210259, url: "https://blog.gopheracademy.com/gophercon-turns-three/", title: "GopherCon Turns Three"},
        ],
      }
    }
  }


  var header = {
    view: function() {
      return m("div", [
          m("div", {className: "navigation"}, [
            a({href: "/ui/add/bookmark", confif: m.route}, ["add bookmark"]),
            m("span", [" | "]),
            a({href: "/ui/add/feed"}, ["add feed"]),
            m("span", [" | "]),
            a({href: "/ui/add/note"}, ["add note"]),
            m("span", [" | "]),
            a({href: "/ui/"}, ["index"]),
          ]),
      ])
    }
  }


  var listing = {
    controller: function() {
      return Entry.list()
    },
    view: function(ctrl) {
      return m("div", [
          header,
          m("div", {className: "content"}, [
            ctrl.entries.map(function(e) {
              return m("div", {className: "entry"}, [
                  m("a", {href: e.url}, [e.title]),
                  m("div", {className: "attrs"}, ["" + Date()]),
              ])
            }),
          ])
      ])
    },
  }

  var addNote = {
    controller: function() {
      var ctrl = {
        content: m.prop(storage.getItem("__new_note__")),
        contentChange: function(e) {
          ctrl.content(e.target.value)
          storage.setItem("__new_note__", e.target.value)
        }
      }
      return ctrl
    },
    view: function(ctrl) {
      return m("div", [
          header,
          m("div", {className: "content"}, [
            m("h1", ["add note"]),
            m("form", [
              m("textarea", {
                required: true,
                value: ctrl.content(),
                onkeypress: ctrl.contentChange,
                onchange: ctrl.contentChange,
              }),
              m("button", {type: 'submit'}, ["save"]),
            ]),
          ]),
      ])
    },
  }

  var addBookmark = {
    controller: function() {
      return {
        url: m.prop(""),
      }
    },
    view: function(ctrl) {
      return m("div", [
          header,
          m("div", {className: "content"}, [
            m("h1", ["add bookmark"]),
            m("form", [
              m("input", {
                required: true,
                type: "url",
                value: ctrl.url(),
                onkeypress: m.withAttr("value", ctrl.url),
                onchange: m.withAttr("value", ctrl.url),
              }),
              m("button", {type: "submit"}, ["add"]),
            ]),
          ]),
      ])
    },
  }

  var addFeed = {
    controller: function() {
      return {
        url: m.prop(""),
      }
    },
    view: function(ctrl) {
      return m("div", [
          header,
          m("div", {className: "content"}, [
            m("h1", ["add feed"]),
            m("form", [
              m("input", {
                required: true,
                type: "url",
                value: ctrl.url(),
                onkeypress: m.withAttr("value", ctrl.url),
                onchange: m.withAttr("value", ctrl.url),
              }),
              m("button", {type: "submit"}, ["add"]),
            ]),
          ]),
      ])
    },
  }

  function a(attrs, content) {
    attrs.onclick = function(e) {
      e.preventDefault()
      m.route(attrs.href)
    }
    return m("a", attrs, content)
  }

  m.route.mode = 'pathname'

  window.onload = function() {
    m.route(document.getElementById("application"), "/ui/", {
      '/ui/': listing,
      '/ui/add/note': addNote,
      '/ui/add/bookmark': addBookmark,
      '/ui/add/feed': addFeed,
    })
  }

}())
