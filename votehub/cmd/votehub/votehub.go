package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/husio/apps/votehub/ghub"
	"github.com/husio/apps/votehub/votes"
	"github.com/husio/apps/votehub/webhooks"
	"github.com/husio/x/auth"
	"github.com/husio/x/cache"
	"github.com/husio/x/envconf"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/tmpl"
	"github.com/husio/x/web"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

var router = web.NewRouter("", web.Routes{
	web.GET(`/`, "", votes.HandleListCounters),
	web.GET(`/counters`, "counters-listing", votes.HandleListCounters),

	web.GET(`/v/{counter-id:\d+}/upvote`, "counters-upvote", votes.HandleClickUpvote),
	web.GET(`/v/{counter-id:\d+}/banner.svg`, "counters-banner-svg", votes.HandleRenderSVGBanner),

	web.GET(`/login`, "login", web.RedirectHandler("/login/basic", http.StatusSeeOther)),
	web.GET(`/login/basic`, "login-basic", auth.LoginHandler("github:basic")),
	web.GET(`/login/repo-owner`, "login-repo-owner", auth.LoginHandler("github:repo-owner")),
	web.GET(`/login/github/success`, "login-github-callback", auth.HandleLoginCallback),

	web.GET(`/webhooks/create`, "webhooks-listing", webhooks.HandleListWebhooks),
	web.POST(`/webhooks/create`, "webhooks-create", webhooks.HandleCreateWebhooks),
	web.POST(`/webhooks/callbacks/issues`, "webhooks-issues-callback", webhooks.HandleIssuesWebhookCallback),

	web.ANY(`.*`, "", handle404),
})

func main() {
	log.SetFlags(log.Lshortfile)

	conf := struct {
		HTTP           string
		GithubKey      string `envconf:",required"`
		GithubSecret   string `envconf:",required"`
		Postgres       string `envconf:",required"`
		Statics        string
		Templates      string
		TemplatesCache bool
		Schema         string
	}{
		HTTP:      "localhost:8000",
		Templates: "**/templates/**.html",
	}
	envconf.Must(envconf.LoadEnv(&conf))

	tmpl.MustLoadTemplates(conf.Templates, conf.TemplatesCache)

	ctx := context.Background()
	ctx = auth.WithOAuth(ctx, map[string]*oauth2.Config{
		"github:basic": &oauth2.Config{
			ClientID:     conf.GithubKey,
			ClientSecret: conf.GithubSecret,
			Scopes:       []string{},
			Endpoint:     oauth2gh.Endpoint,
		},
		"github:repo-owner": &oauth2.Config{
			ClientID:     conf.GithubKey,
			ClientSecret: conf.GithubSecret,
			Scopes:       []string{"public_repo", "write:repo_hook"},
			Endpoint:     oauth2gh.Endpoint,
		},
	})
	ctx = ghub.WithClient(ctx, ghub.StandardClient)
	ctx = web.WithRouter(ctx, router)
	ctx = cache.WithLocalCache(ctx, 1000)

	if db, err := sql.Open("postgres", conf.Postgres); err != nil {
		log.Fatalf("cannot connect to PostgreSQL: %s", err)
	} else {
		if conf.Schema != "" {
			log.Printf("loading database schema: %s", conf.Schema)
			pg.MustLoadSchema(db, conf.Schema)
		}
		ctx = pg.WithDB(ctx, db)
	}

	app := &application{
		ctx: ctx,
		rt:  router,
	}
	log.Printf("running HTTP server: %s", conf.HTTP)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Printf("HTTP server error: %s", err)
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
	fmt.Printf(":: %5s %5s %s\n", workTime, r.Method, r.URL)
}
