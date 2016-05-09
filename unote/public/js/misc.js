/* globals window */
/* globals document */
define(["lib/mithril"], function(m) {

  function stdview(view) {
    return function(ctrl, extra) {
      return m("div", {className: "box"}, [
          m("div", {className: "actions"}, [
            a({href: "/ui/add"}, "create new"),
            a({href: "/ui"}, "list all"),
            a({href: "/#TODO"}, "logout"),
          ]),
          m("div", {className: "content"}, view(ctrl, extra)),
      ])
    }
  }


  function a(attrs, content) {
    attrs.onclick = function(e) {
      e.preventDefault()
      m.route(attrs.href)
    }
    return m("a", attrs, content)
  }


  function textareaHeight() {
    return (window.innerHeight - 38) + 'px';
  }

  return {
    stdview: stdview,
    a: a,
    textareaHeight: textareaHeight,
    window: window,
    document: document,
  }

})
