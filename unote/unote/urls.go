package unote

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/husio/x/web"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

func init() {
	app := application{
		rt: web.NewRouter(web.Routes{
			{"/api/notes", handleListNotes, "GET"},
			{"/api/notes", handleAddNote, "POST"},
			{"/api/notes/{note-id}", handleGetNote, "GET"},
			{"/api/.*", handleApi404, web.AnyMethod},

			{"/", handleIndex, "GET"},
			{"/ui", handleIndex, "GET"},
			{"/login", handleLogin, "GET"},
			{"/.*", handle404, web.AnyMethod},
		}),
	}
	http.Handle("/", &app)
}

type application struct {
	rt *web.Router
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	app.rt.ServeCtxHTTP(ctx, w, r)
}

func handleApi404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	web.StdJSONResp(w, http.StatusNotFound)
}

func handle404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func handleLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
