define(["db", "lib/mithril"], function(db, m) {

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
      var obj = db.getItem("unote:" + noteId)
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
      db.setItem("unote:" + note.noteId(), note)

      var lst = db.getItem("unote:list", [])
      if (!lst.includes(note.noteId())) {
        lst.unshift(note.noteId())
        db.setItem("unote:list", lst)
      }

      d.resolve(note)
      return d.promise
    },
    list: function() {
      var d = m.deferred()
      var notes = db.getItem("unote:list", []).map(function(noteId) {
        return new Notes.model(db.getItem("unote:" + noteId))
      })
      d.resolve(notes)
      return d.promise
    },
    destroy: function(noteId) {
      var d = m.deferred()
      db.removeItem("unote:" + noteId)

      var lst = db.getItem("unote:list", [])
      for (var i=0;i<lst.length;i++) {
        if (lst[i] === noteId) {
          lst.splice(i, 1)
          db.setItem("unote:list", lst)
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
