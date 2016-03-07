package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/husio/apps/webcron/webcron"
)

const (
	storagePath = "state.json"
	httpAddr    = "localhost:8000"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Lshortfile | log.Ltime)

	ctx := context.Background()
	errc := make(chan error, 1)

	storage, err := webcron.NewFileStorage(storagePath)
	if err != nil {
		log.Fatalf("cannot create storage: %s", err)
	}
	log.Printf("storage initialized: %s", storagePath)
	defer storage.Close()

	scheduler, err := webcron.RunScheduler(ctx, storage)
	if err != nil {
		log.Fatalf("cannot create scheduler: %s", err)
	}

	ui := webcron.NewHandler(scheduler)

	go func() {
		log.Printf("HTTP server started: %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, ui); err != nil {
			errc <- fmt.Errorf("HTTP server error: %s", err)
			return
		}
	}()

	go func() {
		for err := range scheduler.Errc() {
			log.Printf("scheduler error: %s", err)
		}
	}()

	if err, ok := <-errc; ok {
		log.Fatal(err)
	}
}
