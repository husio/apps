package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/husio/apps/textscore/article"
)

func main() {
	printArticleFl := flag.Bool("p", false, "Print articles")
	flag.Parse()

	var b bytes.Buffer

	rd := bufio.NewReader(os.Stdin)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		resp, err := http.Get(line)
		if err != nil {
			log.Printf("cannot GET %s: %s", line, err)
		}

		b.Reset()
		err = article.Parse(io.LimitReader(resp.Body, 1000000), &b)
		resp.Body.Close()
		if *printArticleFl {
			b.WriteTo(os.Stdout)
		} else {
			fmt.Printf("%6v %s\n", err == nil, line)
		}
	}
}
