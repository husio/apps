package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"

	"github.com/husio/apps/paste/notes"
	"github.com/husio/x/auth"
	"github.com/husio/x/cache"
	"github.com/husio/x/envconf"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/tmpl"
	"github.com/husio/x/web"
)

var router = web.NewRouter("", web.Routes{
	web.GET("/login", "login", auth.LoginHandler("github")),
	web.GET("/login/success", "", auth.HandleLoginCallback),

	web.GET("/n/{note-id}", "note-details", notes.HandleDisplayNote),

	web.GET(`/static/.*`, "", handleStaticDir),
})

var statics http.Handler

func handleStaticDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	statics.ServeHTTP(w, r)
}

func handleApi404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	web.StdJSONErr(w, http.StatusNotFound)
}

func main() {
	log.SetFlags(log.Lshortfile)

	conf := struct {
		GithubKey      string `envconf:",required"`
		GithubSecret   string `envconf:",required"`
		HTTP           string
		Postgres       string
		Statics        string `envconf:",required"`
		Templates      string `envconf:",required"`
		TemplatesCache bool
		Schema         string
	}{
		HTTP:           "localhost:8000",
		Postgres:       "user=paste password=paste dbname=paste sslmode=disable",
		Schema:         "schema.sql",
		TemplatesCache: false,
	}
	envconf.Must(envconf.LoadEnv(&conf))

	tmpl.MustLoadTemplates(conf.Templates, conf.TemplatesCache)

	ctx := context.Background()

	ctx = auth.WithOAuth(ctx, map[string]*oauth2.Config{
		"github": &oauth2.Config{
			ClientID:     conf.GithubKey,
			ClientSecret: conf.GithubSecret,
			Scopes:       []string{},
			Endpoint:     oauth2gh.Endpoint,
		},
	})

	statics = http.StripPrefix("/static", http.FileServer(http.Dir(conf.Statics)))

	ctx = web.WithRouter(ctx, router)
	ctx = cache.WithLocalCache(ctx, 1000)

	if db, err := sql.Open("postgres", conf.Postgres); err != nil {
		log.Fatalf("cannot connect to PostgreSQL: %s", err)
	} else {
		ctx = pg.WithDB(ctx, db)
		if conf.Schema != "" {
			log.Printf("loading schema from %q", conf.Schema)
			pg.MustLoadSchema(db, conf.Schema)
		}
	}

	app := &application{
		ctx: ctx,
		rt:  router,
	}
	log.Printf("running HTTP server: %s", conf.HTTP)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Fatalf("HTTP server error: %s", err)
	}
}

func handle404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

type application struct {
	ctx context.Context
	rt  *web.Router
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	app.rt.ServeCtxHTTP(app.ctx, w, r)
	workTime := time.Now().Sub(start) / time.Millisecond * time.Millisecond
	fmt.Printf("%.6s %5s %s\n", r.Method, workTime, r.URL)
}
