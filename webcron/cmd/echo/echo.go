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
	log.Printf("%s %s", r.Method, r.URL, body)
	w.Write(body)
}
