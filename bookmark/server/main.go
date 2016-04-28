package main

import (
	"database/sql"
	"net/http"

	"github.com/husio/x/envconf"
	"github.com/husio/x/log"
	"github.com/husio/x/storage/pg"
	"golang.org/x/net/context"
)

func main() {
	conf := struct {
		HTTP     string
		Postgres string
		Schema   string
	}{
		HTTP:     "localhost:8000",
		Postgres: "dbname=postgres user=postgres sslmode=disable",
	}
	if err := envconf.LoadEnv(&conf); err != nil {
		log.Fatal("cannot load configuration", "error", err.Error())
	}

	ctx, done := context.WithCancel(context.Background())
	defer done()

	db, err := sql.Open("postgres", conf.Postgres)
	if err != nil {
		log.Fatal("cannot connect to database", "error", err.Error())
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Error("cannot ping database",
			"postgres", conf.Postgres,
			"error", err.Error())
	}
	ctx = pg.WithDB(ctx, db)

	if conf.Schema != "" {
		if err := pg.LoadSchema(db, conf.Schema); err != nil {
			log.Error("cannot load schema",
				"schema", conf.Schema,
				"error", err.Error())
		}
	}

	app := NewApplication(ctx)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Error("HTTP server error", "error", err.Error())
	}
}
