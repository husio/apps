package main

import (
	"database/sql"
	"net/http"

	"github.com/husio/apps/bb/bb"
	"github.com/husio/x/envconf"
	"github.com/husio/x/log"
	"github.com/husio/x/storage/pg"
	"golang.org/x/net/context"
)

func main() {
	conf := struct {
		HTTP     string
		Postgres string
	}{
		HTTP:     "localhost:8000",
		Postgres: "host=localhost port=5432 user=postgres dbname=postgres sslmode=disable",
	}
	envconf.Must(envconf.LoadEnv(&conf))

	ctx := context.Background()

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
