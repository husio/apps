/* globals document */
define(["lib/mithril"], function(m) {

  function module(name) {
    return {
      controller: function() {
        m.startComputation()
        require([name], function(module) {
          this.controller = new module.controller()
          this.view = module.view
          m.endComputation()
        }.bind(this))
      },
      view: function(ctrl) {
        return ctrl.view(ctrl.controller)
      }
    }
  }

  m.route.mode = 'pathname'

  function init() {
    m.route(document.getElementById("application"), "/ui", {
      '/ui': module("listNotes"),
      '/ui/add': module("addNote"),
      '/ui/:noteId': module("noteDetails"),
    })
  }

  return init

})
