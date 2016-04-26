// jshint asi:true
// jshint unused:true
(function() {
  "use strict"

  var s = window.localStorage
  var storage = {
    dirty: function(v) {
      if (v) {
        s.setItem("__new_note__", v)
      }
      return s.getItem("__new_note__")
    },
    list: function() {
      return JSON.parse(s.getItem("list")) || []
    },
    add: function(v) {
      var list = storage.list()
      list.push(v)
      s.setItem("list", JSON.stringify(list))
    },
  }


  function stdview(view) {
    return function(ctrl, extra) {
      return m("div", {className: "box"}, [
          m("div", {className: "actions"}, [
            a({href: "/ui/add"}, "new"),
            m("span", " | "),
            a({href: "/ui"}, "list"),
          ]),
          m("div", {className: "content"}, view(ctrl, extra)),
      ])
    }
  }

  var listing = {
    controller: function() {
      return {
        notes: storage.list(),
      }
    },
    view: stdview(function(ctrl) {
      return m("div", [
        m("p", "recent notes"),
        ctrl.notes.map(function(n) {
          return m("div", n)
        }),
      ])
    }),
  }

  var addNote = {
    controller: function() {
      var ctrl = {
        content: m.prop(storage.dirty()),
        contentChange: function(e) {
          ctrl.content(e.target.value)
          storage.dirty(e.target.value)
        },
        submit: function(e) {
          e.preventDefault()
          if (ctrl.content().length === 0) {
            return
          }
          storage.add(ctrl.content())
          ctrl.content("")
        },
      }
      return ctrl
    },
    view: stdview(function(ctrl) {
      return m("div", [
          m("div", {className: "content"}, [
            m("p", ["add note"]),
            m("form", [
              m("textarea", {
                required: true,
                value: ctrl.content(),
                onkeypress: ctrl.contentChange,
                onchange: ctrl.contentChange,
              }),
              m("button", {type: 'submit', onclick: ctrl.submit}, ["save"]),
            ]),
          ]),
      ])
    }),
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
    m.route(document.getElementById("application"), "/ui", {
      '/ui': listing,
      '/ui/add': addNote,
    })
  }

}())
