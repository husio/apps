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
      var ctrl = {
        content: m.prop(localdb.getItem("__dirty__") || ""),
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
          m.route("/ui/")
        },
        cleanForm: function(e) {
          e.preventDefault()
          localdb.removeItem("__dirty__")
          ctrl.content("")
        }
      }
      return ctrl
    },
    view: function(ctrl) {
      var cleanForm = m("button", {
        className: "btn ",
        disabled: ctrl.content().length === 0,
        onclick: ctrl.cleanForm
      }, "clean")

      return m("div", {className: "content"},
        m("form", [
          m("div", {className: "toolbar"}, [
            m("span", {className: "btn", onclick: routeTo("/ui")}, "cancel"),
            cleanForm,
            m("button", {className: "btn green pull-right", type: 'submit', onclick: ctrl.submit}, ["save"]),
          ]),
          m("textarea", {
            className: "unote-input",
            placeholder: "Create new note..",
            required: true,
            style: {height: misc.textareaHeight()},
            onkeypress: ctrl.contentChange,
            onchange: ctrl.contentChange,
          }, ctrl.content()),
      ]))
    }
  }

})
