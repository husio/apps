package feedreader

import (
	"fmt"
	"hash/fnv"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/husio/x/log"
	"github.com/mmcdole/gofeed"
)

func fetch(urls []string) Entries {
	var result Entries

	p := gofeed.NewParser()

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Error("cannot fetch feed", "url", url, "error", err.Error())
			continue
		}

		feed, err := p.Parse(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Error("cannot parse feed", "url", url, "error", err.Error())
			continue
		}
		for _, it := range feed.Items {
			result = append(result, &Entry{
				Feed: Feed{
					Title: html.UnescapeString(feed.Title),
					Link:  feed.Link,
				},
				Title:     html.UnescapeString(it.Title),
				Link:      it.Link,
				Published: parseTime(it.Published),
			})
		}
	}
	return result
}

var feeds struct {
	sync.RWMutex
	entries []*Entry
}

type Entry struct {
	Feed      Feed
	Title     string
	Link      string
	Published time.Time
}

type Feed struct {
	Title string
	Link  string
}

type Entries []*Entry

func (e Entries) Len() int           { return len(e) }
func (e Entries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e Entries) Less(i, j int) bool { return e[i].Published.Before(e[j].Published) }

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"favicon":       favicon,
	"hashcolor":     hashcolor,
	"truncatechars": truncatechars,
}).ParseGlob(getenv("TEMPLATES", "*/*/*.tmpl")))

var debugTemplate = os.Getenv("TEMPLATE_DEBUG") == "true"

func render(w io.Writer, name string, context interface{}) {
	if debugTemplate {
		tmpl = template.Must(template.New("").Funcs(template.FuncMap{
			"favicon":       favicon,
			"hashcolor":     hashcolor,
			"truncatechars": truncatechars,
		}).ParseGlob(getenv("TEMPLATES", "*/*/*.tmpl")))
	}
	if err := tmpl.ExecuteTemplate(w, name, context); err != nil {
		log.Error("cannot render template", "name", name, "error", err.Error())
	}
}

func getenv(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func hashcolor(s string) string {
	hash := fnv.New64()
	fmt.Fprint(hash, s)
	i := hash.Sum64()

	r := (i & 0xFF0000) >> 16
	g := (i & 0x00FF00) >> 8
	b := i & 0x0000FF
	return fmt.Sprintf("#%X%X%X", r, g, b)
}

func favicon(link string) string {
	return "//www.google.com/s2/favicons?domain_url=" + url.QueryEscape(link)
}

func truncatechars(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func sources() ([]string, error) {
	b, err := ioutil.ReadFile("sources.txt")
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(b)), nil
}

func Update() {
	urls, err := sources()
	if err != nil {
		log.Error("cannot get sources", "error", err.Error())
		return
	}
	entries := fetch(urls)
	sort.Sort(sort.Reverse(entries))

	feeds.Lock()
	defer feeds.Unlock()

	feeds.entries = entries
}
