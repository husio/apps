// jshint asi:true
// jshint unused:true
define([], function() {

  m.route.mode = 'pathname'

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

  return function() {
    m.route(document.getElementById("application"), "/ui/", {
      "/ui/": module("pictures"),
      "/ui/upload": module("pictureUploaderCategory"),
      "/ui/upload/:categoryName": module("pictureUploader"),
      "/ui/picture/:imageId": module("pictureDetails"),
    })
  }
})
