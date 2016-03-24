package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/husio/apps/auth-service/auth"
	"github.com/husio/x/envconf"
	"github.com/husio/x/stamp"
	"github.com/husio/x/storage/pg"

	_ "github.com/lib/pq"
	"golang.org/x/net/context"
)

func main() {
	conf := struct {
		HTTP     string
		Postgres string
		Secret   []byte `envconf:",required"`
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
	ctx = auth.WithTokenSigner(ctx, stamp.NewHMAC256Signer(conf.Secret))

	app := auth.NewApp(ctx)
	log.Printf("running http server: %s", conf.HTTP)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Printf("HTTP server error: %s", err)
	}
}
