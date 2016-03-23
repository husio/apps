package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/husio/apps/auth-service/auth"
	"github.com/husio/x/stamp"
	"github.com/husio/x/storage/pg"

	"golang.org/x/net/context"
)

// lazy me...
const (
	httpAddr = "localhost:8000"
	pgConf   = "host=localhost port=5432 user=postgres dbname=postgres"
	secret   = "qfwoifwoifhwofihoihqwfofwqihfqohifq"
)

func main() {
	ctx := context.Background()

	db, err := sql.Open("postgres", pgConf)
	if err != nil {
		log.Fatalf("cannot open database: %s", err)
	}
	defer db.Close()

	ctx = pg.WithDB(ctx, db)
	ctx = auth.WithTokenSigner(ctx, stamp.NewHMAC256Signer([]byte(secret)))

	app := auth.NewApp(ctx)
	log.Printf("running http server: %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, app); err != nil {
		log.Printf("HTTP server error: %s", err)
	}
}
