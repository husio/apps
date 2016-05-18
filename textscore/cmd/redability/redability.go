package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/husio/x/log"
	"github.com/mauidude/go-readability"
)

func main() {
	limitFl := flag.Int64("limit", 10240, "Page read size limit")
	flag.Parse()

	for _, urlStr := range os.Args[1:] {
		resp, err := http.Get(urlStr)
		if err != nil {
			log.Fatal("cannot GET url", "url", urlStr, "error", err.Error())
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(io.LimitReader(resp.Body, *limitFl))
		if err != nil {
			log.Fatal("cannot read HTTP response", "url", urlStr, "error", err.Error())
		}

		s := html.UnescapeString(string(b))
		doc, err := readability.NewDocument(s)
		if err != nil {
			log.Fatal("cannot parse page", "error", err.Error())
		}

		fmt.Println(doc.Content())
	}
}
