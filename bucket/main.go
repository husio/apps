package main

import "net/http"

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.ListenAndServe(":8000", nil)
}
