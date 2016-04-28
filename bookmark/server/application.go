package main

import (
	"net/http"

	"github.com/husio/x/web"
	"golang.org/x/net/context"
)

type application struct {
	ctx context.Context
	rt  *web.Router
}

func NewApplication(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{

			{`/api/bookmarks`, handleListBookmarks, "GET"},
			{`/api/bookmarks`, cors(handleAddBookmark), "POST,OPTIONS"},

			{`/static/{path:.*}`, handleStatics("./public"), "GET"}, // TODO
			{`/{path:.*}`, handleStatics("./public"), "GET"},        // TODO

			{`.*`, web.StdJSONHandler(http.StatusNotFound), "GET,POST,PUT,DELETE"},
		}),
	}
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.rt.ServeCtxHTTP(app.ctx, w, r)
}

func cors(handler web.HandlerFunc) web.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		if o := r.Header.Get("Origin"); o != "" {
			h.Set("Access-Control-Allow-Origin", o)
		}
		h.Set("Access-Control-Allow-Methods", "POST, GET")
		h.Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		handler(ctx, w, r)
	}
}

func handleStatics(root string) web.HandlerFunc {
	h := http.FileServer(http.Dir(root))
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		r.URL.Path = web.Args(ctx).ByIndex(0)
		h.ServeHTTP(w, r)
	}
}
