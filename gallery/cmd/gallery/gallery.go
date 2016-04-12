package main

import (
	"database/sql"
	"net/http"

	"golang.org/x/net/context"

	"github.com/husio/apps/gallery/gallery"
	"github.com/husio/x/log"
	"github.com/husio/x/storage/sq"
)

func main() {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "gallery.sqlite3")
	if err != nil {
		log.Fatal("cannot open database", "error", err.Error())
	}
	if err := db.Ping(); err != nil {
		log.Fatal("cannot ping database", "error", err.Error())
	}
	ctx = sq.WithDB(ctx, db)

	app := gallery.NewApplication(ctx)

	log.Debug("running HTTP server", "address", ":8000")
	if err := http.ListenAndServe(":8000", app); err != nil {
		log.Fatal("HTTP server error", "error", err.Error())
	}
}
