package main

import (
	"database/sql"
	"net/http"

	"github.com/husio/apps/bb/bb"
	"github.com/husio/x/auth"
	"github.com/husio/x/cache"
	"github.com/husio/x/envconf"
	"github.com/husio/x/log"
	"github.com/husio/x/storage/pg"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	oauth2google "golang.org/x/oauth2/google"
)

func main() {
	conf := struct {
		HTTP     string
		Postgres string
	}{
		HTTP:     "localhost:8000",
		Postgres: "host=localhost port=5432 user=postgres dbname=bb sslmode=disable",
	}
	envconf.Must(envconf.LoadEnv(&conf))

	ctx := context.Background()

	ctx = auth.WithOAuth(ctx, map[string]*oauth2.Config{
		"google": &oauth2.Config{
			ClientID:     "352914691292-2h70272sb408r3vibe4jm4egote804ka.apps.googleusercontent.com",
			ClientSecret: "L_bgOHLCgNYL-3KG8a5u99mF",
			RedirectURL:  "http://bb.example.com:8000/login/success",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: oauth2google.Endpoint,
		},
	})

	ctx = cache.WithLocalCache(ctx, 1000)

	db, err := sql.Open("postgres", conf.Postgres)
	if err != nil {
		log.Fatal("cannot open database", "error", err.Error())
	}
	defer db.Close()
	ctx = pg.WithDB(ctx, db)
	go func() {
		if err := db.Ping(); err != nil {
			log.Error("cannot ping database", "error", err.Error())
		}
	}()

	app := bb.NewApp(ctx)
	log.Debug("running HTTP server", "address", conf.HTTP)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Error("HTTP server error", "error", err.Error())
	}
}
