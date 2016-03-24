package entries

import (
	"log"
	"net/http"

	"github.com/husio/x/storage/pg"
	"github.com/husio/x/tmpl"

	"golang.org/x/net/context"
)

func HandleListEntries(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	entries, err := ListEntries(pg.DB(ctx), 100, 0)
	if err != nil {
		log.Printf("cannot get entries: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		tmpl.Render(w, "standard_response.html", http.StatusText(http.StatusInternalServerError))
		return
	}

	context := struct {
		Entries []*Entry
	}{
		Entries: entries,
	}
	tmpl.Render(w, "entry_list.html", context)
}

func HandleListResources(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	code := http.StatusNotImplemented
	w.WriteHeader(code)
	tmpl.Render(w, "standard_response.html", http.StatusText(code))
}

func HandleAddResource(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	code := http.StatusNotImplemented
	w.WriteHeader(code)
	tmpl.Render(w, "standard_response.html", http.StatusText(code))
}
