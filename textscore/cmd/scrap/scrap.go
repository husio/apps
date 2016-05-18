package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/garyburd/redigo/redis"
	"github.com/husio/x/log"
)

func main() {
	redisFl := flag.String("redis", "redis://127.0.0.1:6379/0", "Redis address")
	stopwFl := flag.String("stopw", "", "Stopwords list")
	httpFl := flag.String("http", "localhost:8000", "HTTP server address")
	flag.Parse()

	rp := redisPool(*redisFl)
	defer rp.Close()

	urlc := make(chan string, 100)

	stopw := make(map[string]struct{})
	if *stopwFl != "" {
		stopw = stopwords(*stopwFl)
	}

	for i := 0; i < 10; i++ {
		go crawler(rp, urlc, stopw)
	}

	for _, u := range flag.Args() {
		urlc <- u
	}

	handleArticles := articleHandler(rp)
	handleScrap := scrapHandler(urlc)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		switch r.Method {
		case "GET":
			handleArticles(w, r)
		case "POST":
			handleScrap(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	if err := http.ListenAndServe(*httpFl, nil); err != nil {
		log.Fatal("HTTP server error", "error", err.Error())
	}
}

func redisPool(url string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func crawler(rp *redis.Pool, urlc chan string, stopw map[string]struct{}) {
	for urlStr := range urlc {
		for _, u := range crawl(rp, urlStr, stopw) {
			select {
			case urlc <- u:
			default:
				// XXX
				// drop links - no spare worker
				continue
			}
		}
	}
}

func crawl(rp *redis.Pool, urlStr string, stopw map[string]struct{}) []string {
	log.Debug("crawling started", "url", urlStr)
	defer log.Debug("crawling done", "url", urlStr)

	resp, err := httpcli.Get(urlStr)
	if err != nil {
		log.Error("cannot GET",
			"url", urlStr,
			"error", err.Error())
		return nil
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !isHtml(ct) {
		log.Debug("non HTML resource",
			"url", urlStr,
			"contentType", ct)
		return nil
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 100000))
	if err != nil {
		log.Error("cannot read body",
			"url", urlStr,
			"error", err.Error())
		return nil
	}
	text := htmlToText(bytes.NewReader(body))
	key := fmt.Sprintf("article:%x", sha1.Sum(text))

	rc := rp.Get()
	defer rc.Close()

	if exists, err := redis.Bool(rc.Do("EXISTS", key)); err != nil {
		log.Error("cannot query database",
			"url", urlStr,
			"error", err.Error())
		return nil
	} else if exists {
		log.Debug("article already stored",
			"url", urlStr,
			"key", key)
		return nil
	}

	func() {
		// all redis update is done in single batch
		if err := rc.Send("MULTI"); err != nil {
			log.Error("cannot start MULTI", "error", err.Error())
			return
		}

		err = rc.Send("HMSET", key,
			"url", urlStr,
			"title", pageTitle(body),
			"crated", time.Now().Unix())
		if err != nil {
			log.Error("cannot write article data",
				"key", key,
				"url", urlStr,
				"error", err.Error())
			return
		}

		for w, n := range words(bytes.NewReader(text), stopw) {
			if len(w) < 3 {
				continue
			}
			if err := rc.Send("ZADD", "word:"+w, n, key); err != nil {
				log.Error("cannot write word count",
					"key", key,
					"url", urlStr,
					"word", w,
					"error", err.Error())
				return
			}
		}
		if _, err := rc.Do("EXEC"); err != nil {
			log.Error("cannot flush redis command",
				"key", key,
				"url", urlStr,
				"error", err.Error())
			return
		}
	}()

	return pageUrls(body)
}

var httpcli = http.Client{
	Timeout: 10 * time.Second,
}

const oneMB = 1000000

func pageUrls(body []byte) []string {
	var urls []string

	t := html.NewTokenizer(bytes.NewReader(body))

	for {
		switch token := t.Next(); token {
		case html.StartTagToken:
			if tt := t.Token(); tt.Data == "a" {
				for _, a := range tt.Attr {
					if a.Key == "href" {
						urls = append(urls, a.Val)
					}
				}
			}
		case html.ErrorToken:
			return urls
		}
	}
}

func isHtml(contentType string) bool {
	for _, ct := range strings.Split(contentType, ";") {
		if strings.TrimSpace(ct) == "text/html" {
			return true
		}
	}
	return false
}

var ignoretag map[string]struct{} = map[string]struct{}{
	"script": struct{}{},
	"code":   struct{}{},
	"iframe": struct{}{},
}

func htmlToText(r io.Reader) []byte {
	t := html.NewTokenizer(r)

	var out bytes.Buffer

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
			return out.Bytes()
		case html.CommentToken:
			continue
		case html.TextToken:
			if ignorescore == 0 {
				html := strings.TrimSpace(t.Token().Data)
				if len(html) > 0 {
					fmt.Fprintln(&out, html)
				}
			}
		}
	}
}

func words(r io.Reader, stopw map[string]struct{}) map[string]int {
	counts := make(map[string]int)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		w := strings.ToLower(scanner.Text())
		if strings.HasPrefix(w, "<") || strings.HasSuffix(w, ">") {
			continue
		}
		w = strings.TrimRight(w, ",.")

		if len(w) > 40 {
			continue
		}

		if _, ok := stopw[w]; ok {
			continue
		}

		counts[w]++
	}

	if err := scanner.Err(); err != nil {
		log.Error("scanner error", "error", err.Error())
	}
	return counts
}

func stopwords(path string) map[string]struct{} {
	stopw := make(map[string]struct{})
	fd, err := os.Open(path)
	if err != nil {
		log.Error("cannot open stopwords file", "error", err.Error())
		return stopw
	}
	defer fd.Close()

	rd := bufio.NewReader(fd)
	for {
		word, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Error("cannot read stopwords", "error", err.Error())
			}
			return stopw
		}
		stopw[strings.TrimSpace(word)] = struct{}{}
	}
}

func pageTitle(body []byte) string {
	match := matchTitle(body, 1)
	if len(match) == 0 {
		return ""
	}
	return html.UnescapeString(string(match[0][1]))
}

var matchTitle = regexp.MustCompile("<title>(.*?)</title>").FindAllSubmatch

func scrapHandler(urlc chan string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("cannot read body", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go func() {
			for _, urlStr := range strings.Fields(string(b)) {
				urlc <- urlStr
			}
		}()

		fmt.Fprintln(w, "ok")
	}
}

func articleHandler(rp *redis.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Article struct {
			Key   string `redis:"-"`
			Url   string `redis:"url"`
			Title string `redis:"title"`
		}

		rc := rp.Get()
		defer rc.Close()

		var articles []*Article
		for _, w := range r.URL.Query()["word"] {
			keys, err := redis.Strings(rc.Do("ZREVRANGE", "word:"+w, 0, 100))
			if err != nil {
				log.Error("cannot get keys", "error", err.Error())
				continue
			}
			for _, key := range keys {
				raw, err := redis.Values(rc.Do("HGETALL", key))
				if err != nil {
					log.Error("cannot get article",
						"key", key,
						"error", err.Error())
					continue
				}
				var art Article
				if err := redis.ScanStruct(raw, &art); err != nil {
					log.Error("cannot scan article",
						"key", key,
						"error", err.Error())
					continue
				}
				art.Key = key
				articles = append(articles, &art)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(articles)
	}
}
