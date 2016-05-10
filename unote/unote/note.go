package unote

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type Note struct {
	NoteID  string    `json:"noteId"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
}

func handleListNotes(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	notes, err := ListNotes(ctx)
	if err != nil {
		log.Printf("cannot read note: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Notes []*Note `json:"notes"`
	}{
		Notes: notes,
	}
	JSONResp(w, resp, http.StatusOK)
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

func handleAddNote(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string    `json:"content"`
		Created time.Time `json:"created"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		JSONErr(w, err.Error(), http.StatusBadRequest)
		return
	}

	var errs []string
	if input.Content == "" {
		errs = append(errs, `"content" is required`)
	}
	if len(errs) != 0 {
		JSONErrs(w, errs, http.StatusBadRequest)
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

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "Note", n.NoteID, 0, nil)
	_, err := datastore.Put(ctx, key, &n)
	if err != nil {
		log.Printf("cannot put note: %s", err)
		StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	JSONResp(w, &n, http.StatusCreated)
}

func handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	StdJSONResp(w, http.StatusNotImplemented)
}
