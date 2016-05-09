/* globals localStorage */
define([], function() {

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
