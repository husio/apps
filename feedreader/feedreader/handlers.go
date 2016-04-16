package feedreader

import (
	"net/http"

	"golang.org/x/net/context"
)

func HandleListEntries(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	feeds.RLock()
	defer feeds.RUnlock()

	context := struct {
		Entries Entries
	}{
		Entries: feeds.entries,
	}
	render(w, "listing.tmpl", context)
}

func HandleListSources(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	sources, _ := sources()
	context := struct {
		Sources []string
	}{
		Sources: sources,
	}
	render(w, "sources.tmpl", context)
}

func Handle404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(http.StatusNotFound)
	content := struct {
		Title string
	}{
		Title: http.StatusText(http.StatusNotFound),
	}
	render(w, "error.tmpl", content)
}
