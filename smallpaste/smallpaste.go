package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"golang.org/x/net/context"

	"github.com/husio/x/log"
	"github.com/husio/x/web"
)

func main() {
	if err := http.ListenAndServe(":8000", rt); err != nil {
		log.Error("HTTP server error", "error", err.Error())
	}
}

var rt = web.NewRouter(web.Routes{
	{"GET", `/{id}`, handleGetPaste},
	{"POST", `/`, handlePostPaste},
	{"PUT", `/{id}`, handlePutPaste},
	{web.AnyMethod, `.*`, handle404},
})

var db = struct {
	mu  sync.Mutex
	mem map[string][]byte
}{
	mem: make(map[string][]byte),
}

func handle404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "not found")
}

func handleGetPaste(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pid := web.Args(ctx).ByIndex(0)

	db.mu.Lock()
	b, ok := db.mem[pid]
	db.mu.Unlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "not found")
	} else {
		w.Write(b)
	}
}

func handlePostPaste(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pid := randomID(16)
	handleStorePaste(pid, w, r)
}

func randomID(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := base64.URLEncoding.EncodeToString(b)
	return s[:size]
}

func handlePutPaste(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pid := web.Args(ctx).ByIndex(0)
	handleStorePaste(pid, w, r)
}

func handleStorePaste(pid string, w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		log.Error("cannot read body", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "internal server error")
		return
	}

	db.mu.Lock()
	db.mem[pid] = b
	defer db.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, pid)
}
