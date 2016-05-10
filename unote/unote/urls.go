package unote

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

func init() {
	http.HandleFunc("/login", handleLogin)

	http.HandleFunc("/api/notes", handleListNotes)

	http.HandleFunc("/", handleIndex)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	next := r.URL.Query().Get("next")
	if next == "" {
		next = r.URL.Query().Get("continue")
	}
	if next == "" {
		next = r.Referer()
	}

	if user.Current(ctx) != nil {
		url, _ := user.LoginURL(ctx, "/")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, next, http.StatusTemporaryRedirect)
	}
}
