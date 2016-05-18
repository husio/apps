package main

import (
	"net/http"
	"time"

	"github.com/husio/apps/auth/keys"
	"github.com/husio/x/envconf"
	"github.com/husio/x/log"
	"golang.org/x/net/context"
)

func main() {
	conf := struct {
		HTTP string
	}{
		HTTP: "localhost:8000",
	}
	if err := envconf.LoadEnv(&conf); err != nil {
		log.Fatal("cannot load configuration", "error", err.Error())
	}

	ctx := context.Background()

	km, stop := keyManager()
	defer stop()
	ctx = keys.WithManager(ctx, km)

	app := NewApplication(ctx)
	if err := http.ListenAndServe(conf.HTTP, app); err != nil {
		log.Error("HTTP server error", "error", err.Error())
	}
}

func keyManager() (*keys.KeyManager, func()) {
	var m keys.KeyManager
	t := time.NewTicker(24 * time.Hour)

	go func() {
		for range t.C {
			if id, err := m.GenerateKey(24 * 7 * time.Hour); err != nil {
				log.Error("cannot generate key", "error", err.Error())
			} else {
				log.Debug("new key generated", "id", id)
			}
		}
	}()

	return &m, t.Stop
}
