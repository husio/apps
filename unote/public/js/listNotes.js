define(["Notes", "misc", "lib/mithril"], function(Notes, misc, m) {

  function routeTo(url) {
    return function(e) {
      e.preventDefault()
      m.route(url)
    }
  }

  return {
    controller: function() {
      return {
        notes: Notes.list(),
        createNote: function(e) {
          e.preventDefault()
          m.route("/ui/add")
        }
      }
    },
    view: function(ctrl) {
      var notes
      if (ctrl.notes().length > 0) {
        notes = ctrl.notes().map(function(n) {
          return m("div", {
            className: "entry link",
            onclick: routeTo("/ui/" + n.noteId()),
          }, n.content().substring(0, 80))
        })
      }

      return m("div", {className: "content"}, [
        m("div", {className: "listing"}, [
            m("div", {
              className: "entry create link",
              onclick: ctrl.createNote,
            }, "Create new"),
            notes,
        ]),
      ])
    },
  }


})
