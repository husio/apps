package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

var ignoretag map[string]struct{} = map[string]struct{}{
	"script": struct{}{},
	"code":   struct{}{},
	"iframe": struct{}{},
}

func main() {
	t := html.NewTokenizer(os.Stdin)

	var ignorescore int
	for {
		switch token := t.Next(); token {
		case html.StartTagToken:
			if _, ok := ignoretag[string(t.Token().Data)]; ok {
				ignorescore++
			}
		case html.EndTagToken:
			if _, ok := ignoretag[string(t.Token().Data)]; ok {
				ignorescore--
			}
		case html.ErrorToken:
			return
		case html.CommentToken:
			continue
		case html.TextToken:
			if ignorescore == 0 {
				html := strings.TrimSpace(t.Token().Data)
				if len(html) > 0 {
					fmt.Print(html)
					fmt.Print(" ")
				}
			}
		}
	}
}
