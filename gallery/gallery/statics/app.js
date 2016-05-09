(function() {
  // jshint asi:true
  // jshint unused:true
  "use strict"




  var pictureUploaderCategory = {
    controller: function() {
      var ctrl = {
        categoryName: m.prop(m.route.param("categoryName") || ""),
        updateCategoryName: function(e) {
          // enter or button press
          if (e.keyCode === 13 || e.keyCode === undefined) {
            e.preventDefault()
            m.route("/ui/upload/" + ctrl.categoryName())
            return
          }

          // normal input
          ctrl.categoryName(e.target.value)
        },
      }
      return ctrl
    },
    view: function(ctrl) {
      return m("div", [
          m("input", {
            value: ctrl.categoryName(),
            placeholder: "Category name",
            onkeyup: ctrl.updateCategoryName,
          }),
          m("button", {onclick: ctrl.updateCategoryName}, "ok"),
      ])
    }
  }

  var pictureUploader = {
    controller: function() {
      var cname = m.route.param("categoryName")
      if (!cname) {
        m.route("/ui/upload")
        return
      }

      var ctrl = {
        categoryName: cname,
      }
      return ctrl
    },
    view: function(ctrl) {
      return m("div", [
          m("h1", "Upload photos to category ", m("em", ctrl.categoryName)),
      ])
    }
  }



  var fileUploader = {
    controller: function () {
      return {
        hover: m.prop(false),
      }
    },
    view: function(ctrl) {
      return m("div", {
        ondragover: function() {
          ctrl.hover(true)
        },
        ondragend: function() {
          ctrl.hover(false)
        },
        ondrop: function(e) {
          e.preventDefault()
          ctrl.hover(false)
          uploadFiles(e.dataTransfer.files)
        },
      }, [
      ])
    },
  }

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


  function uploadFiles(toUpload, ondone) {
    var files = Array.prototype.slice.call(toUpload)

    function uploadNext() {
      var file = files.shift()
      if (file === undefined) {
        return
      }
      uploadFile(file, function (resp, xhr) {
        ondone(file, xhr.status)
        uploadNext()
      })
    }

    uploadNext()
  }

  function uploadFile(file, onready) {
    var data = new FormData()
    data.append('file', file)
    var xhr = new XMLHttpRequest()
    xhr.responseType = 'json'
    xhr.onload = function () {
      if (onready) onready(this.response, xhr)
    }
    xhr.open('PUT', '/api/v1/images') // XXX
    xhr.send(data)
  }





  m.route.mode = 'pathname'

  window.onload = function() {
    m.route(document.getElementById("application"), "/ui/", {
      "/ui/": pictures,
      "/ui/upload": pictureUploaderCategory,
      "/ui/upload/:categoryName": pictureUploader,
      "/ui/picture/:imageId": pictureDetails,
    })
  }

}());
