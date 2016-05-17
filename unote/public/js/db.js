/* globals localStorage */
define(["lib/mithril"], function(m) {

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

  var queue = {
    stack: function() {
      return localdb.getItem("_queue")
    },
    append: function(action) {
      var stack = queue.stack()
      stack.push(action)
      localdb.setItem(stack)
    },
    sync: function() {
      var stack = queue.stack()
      stack.forEach(function(action) {

      })
    }
  }




  return {
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

})
