package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

const httpAddr = "localhost:12345"

func main() {
	http.HandleFunc("/", echo)
	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatalf("HTTP server error: %s", err)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("%s %20q %q", r.Method, r.URL, string(body))
	w.Write(body)
}
