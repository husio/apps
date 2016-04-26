package main

import (
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	public := anyof(os.Getenv("PUBLIC"), "./public")
	httpaddr := anyof(os.Getenv("HTTP"), "localhost:8000")

	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(public))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(public, "index.html"))
	})
	http.ListenAndServe(httpaddr, nil)
}

func anyof(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}
