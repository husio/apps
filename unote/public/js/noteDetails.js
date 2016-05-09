define(["Notes", "localdb", "misc", "lib/mithril"],
    function(Notes, localdb, misc, m) {

  function routeTo(url) {
    return function(e) {
      e.preventDefault()
      m.route(url)
    }
  }

  return {
    controller: function() {
      var raw = localdb.getItem("unote:" + m.route.param("noteId"), null)
      if (raw === null) {
        m.route("/ui")
        return
      }

      var note = new Notes.model(raw)
      var ctrl = {
        content: m.prop(note.content()),
        note: note,
        contentChange: function(e) {
          ctrl.content(e.target.value)
        },
        cleanForm: function(e) {
          e.preventDefault()
          ctrl.content(note.content())
        },
        update: function(e) {
          e.preventDefault()
          note.content(ctrl.content())
          note.save().then(function() {
            m.route("/ui/")
          })
        },
        deleteEntry: function(e) {
          e.preventDefault()
          note.destroy()
          m.route("/ui")
        },
      }
      return ctrl
    },
    view: function(ctrl) {
      var hasChange = ctrl.content() !== ctrl.note.content()

      var backBtn, undoBtn

      if (hasChange) {
        undoBtn = m("button", {
          className: "btn",
          disabled: !hasChange,
          onclick: ctrl.cleanForm,
        }, "undo")
      } else {
        backBtn = m("span", {
          className: "btn",
          onclick: routeTo("/ui"),
        }, "back")
      }

      return m("div", {className: "content"},
        m("form", [
          m("div", {className: "toolbar"}, [
            backBtn,
            undoBtn,
            m("button", {
              className: "btn green pull-right",
              type: 'submit',
              onclick: ctrl.update,
              disabled: !hasChange,
            }, "update"),
            m("button", {
              className: "btn red pull-right",
              onclick: ctrl.deleteEntry,
              disabled: hasChange,
            }, "delete"),
          ]),
          m("textarea", {
            className: "unote-input",
            placeholder: "Create new note..",
            required: true,
            style: {height: misc.textareaHeight()},
            onkeyup: ctrl.contentChange,
            onchange: ctrl.contentChange,
          }, ctrl.content()),
        ])
      )
    },
  }

})
