package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/husio/apps/feedreader/feedreader"
	"github.com/husio/x/log"
	"github.com/husio/x/web"
)

func main() {
	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGUSR1, syscall.SIGHUP)
		defer signal.Stop(sigc)

		log.Debug("updating feeds", "trigger", "init")
		feedreader.Update()

		for {
			select {
			case now := <-time.After(30 * time.Minute):
				log.Debug("updating feeds", "trigger", "clock", "time", now.String())
			case sig := <-sigc:
				log.Debug("updating feeds", "trigger", "signal", "signal", sig.String())
			}
			feedreader.Update()
		}
	}()

	rt := web.NewRouter(web.Routes{
		{"GET", `/`, feedreader.HandleListEntries},
		{"GET", `/sources`, feedreader.HandleListSources},
		{web.AnyMethod, `.*`, feedreader.Handle404},
	})
	if err := http.ListenAndServe(":8000", rt); err != nil {
		log.Error("HTTP server failed", "error", err.Error())
	}
}
