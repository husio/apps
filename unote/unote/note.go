package unote

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/husio/x/log"
	"github.com/husio/x/web"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type Note struct {
	NoteID  string    `json:"noteId"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
}

func handleListNotes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	notes, err := ListNotes(ctx)
	if err != nil {
		log.Error("cannot read note", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Notes []*Note `json:"notes"`
	}{
		Notes: notes,
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func ListNotes(ctx context.Context) ([]*Note, error) {
	q := datastore.NewQuery("Note") //.Filter("User =", u)

	// prevent null JSON response
	notes := make([]*Note, 0, 32)

	t := q.Run(ctx)
	for {
		var n Note
		switch _, err := t.Next(&n); err {
		case nil:
			notes = append(notes, &n)
		case datastore.Done:
			return notes, nil
		default:
			return notes, err
		}
	}
}

func handleAddNote(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string    `json:"content"`
		Created time.Time `json:"created"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.JSONErr(w, err.Error(), http.StatusBadRequest)
		return
	}

	var errs []string
	if input.Content == "" {
		errs = append(errs, `"content" is required`)
	}
	if len(errs) != 0 {
		web.JSONErrs(w, errs, http.StatusBadRequest)
		return
	}

	if input.Created.IsZero() {
		input.Created = time.Now()
	}

	n := Note{
		NoteID:  generateId(),
		Content: input.Content,
		Created: input.Created,
	}

	key := datastore.NewKey(ctx, "Note", n.NoteID, 0, nil)
	_, err := datastore.Put(ctx, key, &n)
	if err != nil {
		log.Debug("cannot put note", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	web.JSONResp(w, &n, http.StatusCreated)
}

func handleGetNote(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var note Note
	key := datastore.NewKey(ctx, "Note", web.Args(ctx).ByIndex(0), 0, nil)
	if err := datastore.Get(ctx, key, &note); err != nil {
		log.Debug("cannot get note",
			"noteId", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		// XXX - what about not found?
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	web.JSONResp(w, note, http.StatusOK)
}
