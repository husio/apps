define(["localdb", "lib/mithril"], function(localdb, m) {

  var Notes = {
    model: function(a) {
      this.noteId = m.prop(a.noteId || randomid(64))
      this.content = m.prop(a.content)
      this.created = m.prop(new Date(a.created ? a.created : new Date()))

      this.save = Notes.save.bind(this, this)
      this.destroy = function() {
        Notes.destroy(this.noteId())
        this.noteId(undefined)
      }
    },
    getById: function(noteId) {
      var d = m.deferred()
      var obj = localdb.getItem("unote:" + noteId)
      if (obj) {
        d.resolve(obj)
      } else {
        d.reject()
      }
      return d.promise
    },
    save: function(note) {
      var d = m.deferred()
      if (!note.noteId()) {
        note.noteId(randomid(64))
      }
      localdb.setItem("unote:" + note.noteId(), note)

      var lst = localdb.getItem("unote:list", [])
      if (!lst.includes(note.noteId())) {
        lst.unshift(note.noteId())
        localdb.setItem("unote:list", lst)
      }

      d.resolve(note)
      return d.promise
    },
    list: function() {
      var d = m.deferred()
      var notes = localdb.getItem("unote:list", []).map(function(noteId) {
        return new Notes.model(localdb.getItem("unote:" + noteId))
      })
      d.resolve(notes)
      return d.promise
    },
    destroy: function(noteId) {
      var d = m.deferred()
      localdb.removeItem("unote:" + noteId)

      var lst = localdb.getItem("unote:list", [])
      for (var i=0;i<lst.length;i++) {
        if (lst[i] === noteId) {
          lst.splice(i, 1)
          localdb.setItem("unote:list", lst)
          break
        }
      }

      d.resolve()
      return d.promise
    },
  }

  function randomid(length) {
        return Math.round((Math.pow(36, length + 1) - Math.random() * Math.pow(36, length))).toString(36).slice(1);
  }


  return Notes

})
