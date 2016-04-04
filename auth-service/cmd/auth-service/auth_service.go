package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/husio/apps/auth-service/auth"
	"github.com/husio/x/envconf"
	"github.com/husio/x/storage/pg"

	_ "github.com/lib/pq"
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
		log.Fatalf("cannot open database: %s", err)
	}
	defer db.Close()
	ctx = pg.WithDB(ctx, db)
	go func() {
		if err := db.Ping(); err != nil {
			log.Printf("cannot ping database: %s", err)
		}
	}()

	var keys auth.KeyManager
	ctx = auth.WithKeyManager(ctx, &keys)
	go func() {
		if err := keys.GenerateKey(3 * 24 * time.Hour); err != nil {
			log.Printf("cannot generate new key: %s", err)
		}
	}()

	app := auth.NewApp(ctx)
	log.Printf("running http server: %s", conf.HTTP)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Printf("HTTP server error: %s", err)
	}
}
