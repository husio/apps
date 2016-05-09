// jshint asi:true
// jshint unused:true
define([], function() {
  var search = {
    controller: function() {
      var ctrl = {
        searchTerm: m.prop(""),
        applyFilter: function() {
          console.log("filter!")
        },
      }
      return ctrl
    },
    view: function(ctrl) {
      return m("div", [
          m("input", {
            onkeyup: function(e) {
              ctrl.searchTerm(e.target.value)
              if (e.keyCode === 13) {
                ctrl.applyFilter()
              }
            },
            value: ctrl.searchTerm(),
            id: "search",
            placeholder: "Search using name=value pairs for filtering",
          }),
      ])
    },
  }

  var pictures = {
    model: function(attrs) {
      this.selected = m.prop(false)
      this.imageId = m.prop(attrs.imageId)
      this.width = m.prop(attrs.width)
      this.height = m.prop(attrs.height)
      this.orientation = m.prop(attrs.orientation)
      this.tags = m.prop(attrs.tags || [])
      this.created = m.prop(new Date(attrs.created))

      var that = this
      this.src = function(query) {
        var url = "/api/v1/images/" + that.imageId() + ".jpg"
        if (!query) {
          return url
        }
        var pairs = []
        for (var k in query) {
          pairs.push(encodeURIComponent(k) + '=' + encodeURIComponent(query[k]))
        }
        return url + '?' + pairs.join('&')
      }
    },
    fetch: function(offset) {
      var off = offset || 0
      return m.request({method: "GET", url: "/api/v1/images?offset=" + off}).then(function(r) {
        return r.images.map(function (attrs) { return new pictures.model(attrs) })
      })
    },
    controller: function() {
      var ps = pictures.fetch(0)
      var ctrl = {
        search: search,
        pictures: ps,
        selected: function() {
          return ps().filter(function (p) { return p.selected() })
        },
        toggleAllSelected: function() {
          ps().forEach(function(p) { p.selected(!p.selected()) })
        },
        toggleSelected: function(p) {
          p.selected(!p.selected())
          ctrl.tagUI.display(ctrl.selected().length !== 0)
        },
        tagUI: {
          tag: m.prop(""),
          display: m.prop(false),
          loading: m.prop(false),
        },
        tagSelected: function(e) {
          e.preventDefault()
          ctrl.tagUI.loading(true)
          var selected = ctrl.selected()
          var sep = ctrl.tagUI.tag().indexOf("=")
          var data = {
            name: ctrl.tagUI.tag().slice(0,sep),
            value: ctrl.tagUI.tag().slice(sep + 1),
          }
          var waiting = selected.length
          selected.forEach(function(p) {
            m.request({
              method: "PUT",
              url: "/api/v1/images/" + p.imageId() + "/tags",
              data: data,
            }).then(function() {
              waiting--
              if (waiting === 0) {
                ctrl.tagUI.loading(false)
              }
            })
          })
        },
        gotoUploader: function(e) {
          e.preventDefault()
          m.route("/ui/upload")
        },
      }
      return ctrl
    },
    view: function(ctrl) {
      var tagUI = null
      if (ctrl.tagUI.display()) {
        tagUI = m("div", [
            m("span", ["tag selected pictures"]),
            m("input", {
              placeholder: "Write tag as name=value pair",
              required: true,
              onchange: m.withAttr("value", ctrl.tagUI.tag),
              disabled: ctrl.tagUI.loading(),
              value: ctrl.tagUI.tag(),
            }),
            m("button", {
              onclick: ctrl.tagSelected,
              disabled: ctrl.tagUI.loading(),
            }, ["save"]),
        ])
      }

      return m("div", [
          ctrl.search,
          m("div", {id: "images"}, [
            m("div", [
              m("button", {onclick: ctrl.toggleAllSelected}, ["toggle selected"]),
              m("button", {onclick: ctrl.gotoUploader}, ["upload new photos"]),
              tagUI,
            ]),
            ctrl.pictures().map(function(p) {
              var className = "picture " + (p.selected() ? "selected" : "")
                return m("img", {
                  className: className,
                  src: p.src({resize: '200x140'}),
                  onclick: function(e) {
                    e.preventDefault()
                      ctrl.toggleSelected(p)
                  },
                })
              //return a({href: "/ui/picture/" + p.imageId()}, [
              //  m("img", {className: className, src: p.src({resize: '200x140'})}),
              //])
            })
          ]),
        ])
    },
  }

  return pictures
})
