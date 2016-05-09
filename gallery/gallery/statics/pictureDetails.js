// jshint asi:true
// jshint unused:true
define([], function() {

  var pictureDetails = {
    controller: function() {
      return {
        picture: new pictures.model({
          imageId: m.route.param("imageId"),
        }),
      }
    },
    view: function(ctrl) {
      return m("div", [
        m("img", {className: "full-size", src: ctrl.picture.src()}),
        fileUploader,
      ])
    },
  }

  return pictureDetails
})
