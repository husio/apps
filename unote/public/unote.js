// jshint asi:true
// jshint unused:true
(function() {
  "use strict"

  var localdb = {
    setItem: function(key, value) {
      localStorage.setItem(key, JSON.stringify(value))
    },
    getItem: function(key, fallback) {
      var it = localStorage.getItem(key)
      if (!it) {
        return fallback
      }
      return JSON.parse(it)
    },
    removeItem: function(key) {
      localStorage.removeItem(key)
    }
  }

  var blink = {
    messages: [],
    add: function(m) {
      blink.messages.push(m)
      return blink
    },
    consume: function() {
      var msgs = blink.messages
      blink.messages = []
      return msgs
    }
  }

  var Notes = {
    model: function(a) {
      this.noteId = m.prop(a.noteId || randomid())
      this.content = m.prop(a.content)
      this.created = m.prop(new Date(a.created ? a.created : new Date()))

      this.save = Notes.save.bind(this, this)
    },
    getById: function(noteId) {
      var d = m.deferred()
      var obj = localdb.getItem("unote:" + noteId)
      if (obj) {
        d.resolve(obj)
      } else {
        d.reject()
      }
      return d.promise
    },
    save: function(note) {
      var d = m.deferred()
      localdb.setItem("unote:" + note.noteId(), note)

      var lst = localdb.getItem("unote:list", [])
      if (!lst.includes(note.noteId())) {
        lst.unshift(note.noteId())
        localdb.setItem("unote:list", lst)
      }

      d.resolve(note)
      return d.promise
    },
    list: function() {
      var d = m.deferred()
      var notes = localdb.getItem("unote:list", []).map(function(noteId) {
        return new Notes.model(localdb.getItem("unote:" + noteId))
      })
      d.resolve(notes)
      return d.promise
    }
  }

  function randomid() {
    return Math.random().toString(36).substring(2) + Math.random().toString(36).substring(2)
  }


  function stdview(view) {
    return function(ctrl, extra) {
      return m("div", {className: "box"}, [
          m("div", {className: "actions"}, [
            a({href: "/ui/add"}, "create new"),
            m("span", "|"),
            a({href: "/ui"}, "list all"),
            m("span", "|"),
            a({href: "/#TODO"}, "logout"),
          ]),
          m("div", {className: "content"}, view(ctrl, extra)),
      ])
    }
  }

  var listNotes = {
    controller: function() {
      return {
        notes: Notes.list(),
      }
    },
    view: stdview(function(ctrl) {
      var notes
      if (ctrl.notes().length === 0) {
        notes = "no notes"
      } else {
        notes = ctrl.notes().map(function(n) {
          return m("div", {className: "entry"}, [
              a({href: "/ui/" + n.noteId()}, n.content().substring(0, 80)),
          ])
        })
      }

      return m("div", {className: "listing"}, [
          notes,
      ])
    }),
  }

  var addNote = {
    controller: function() {
      var ctrl = {
        content: m.prop(localdb.getItem("__dirty__")),
        contentChange: function(e) {
          ctrl.content(e.target.value)
          localdb.setItem("__dirty__", e.target.value)
        },
        submit: function(e) {
          e.preventDefault()
          if (ctrl.content().length === 0) {
            return
          }
          var note = new Notes.model({
            content: ctrl.content(),
            created: new Date(),
          })
          note.save().then(function() {
            ctrl.content("")
            localdb.removeItem("__dirty__")
          })
          m.route("/ui/" + note.noteId())
        },
        cleanForm: function(e) {
          e.preventDefault()
          localdb.removeItem("__dirty__")
          ctrl.content("")
        }
      }
      return ctrl
    },
    view: stdview(function(ctrl) {
      return m("form", [
          m("textarea", {
            className: "unote-input",
            placeholder: "Create new note..",
            required: true,
            style: {height: textareaHeight()},
            onkeypress: ctrl.contentChange,
            onchange: ctrl.contentChange,
          }, ctrl.content()),
          m("div", {className: "unote-input-actions"}, [
            m("span", {className: "link", onclick: ctrl.cleanForm}, ["clean form"]),
            m("button", {type: 'submit', onclick: ctrl.submit}, ["save"]),
          ]),
      ])
    }),
  }

  var noteDetails = {
    controller: function() {
      var note = new Notes.model(localdb.getItem("unote:" + m.route.param("noteId")))
      var ctrl = {
        content: m.prop(note.content()),
        note: note,
        contentChange: function(e) {
          ctrl.content(e.target.value)
        },
        cleanForm: function(e) {
          e.preventDefault()
          ctrl.content(note.content())
        }
      }
      return ctrl
    },
    view: stdview(function(ctrl) {
      var noChange = ctrl.content() === ctrl.note.content()
      return m("form", [
          m("textarea", {
            className: "unote-input",
            placeholder: "Create new note..",
            required: true,
            style: {height: textareaHeight()},
            onkeypress: ctrl.contentChange,
            onchange: ctrl.contentChange,
          }, ctrl.content()),
          m("div", {className: "unote-input-actions"}, [
            m("span", {className: "link", onclick: ctrl.cleanForm}, ["revert changes"]),
            m("span", {className: "link", onclick: ctrl.deleteEntry}, ["delete note"]),
            m("button", {type: 'submit', onclick: ctrl.submit, disabled: noChange}, ["update"]),
          ]),
      ])
    }),
  }

  function textareaHeight() {
    return window.innerHeight - 140 + 'px';
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
      '/ui': listNotes,
      '/ui/add': addNote,
      '/ui/:noteId': noteDetails,
    })
  }

}())
