package main

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/husio/x/log"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/web"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

type Bookmark struct {
	BookmarkID int64     `db:"bookmark_id" json:"id"`
	Title      string    `db:"title"       json:"title"`
	Url        string    `db:"url"         json:"url"`
	Created    time.Time `db:"created"     json:"created"`
}

func handleListBookmarks(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if offset < 0 {
		offset = 0
	}

	bookmarks := make([]*Bookmark, 0, 100)
	err := pg.DB(ctx).Select(&bookmarks, `
		SELECT b.*
		FROM bookmarks b
		ORDER BY created DESC
		LIMIT $1 OFFSET $2
	`, 500, offset)
	if err != nil {
		log.Error("cannot select bookmarks", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	resp := struct {
		Bookmarks []*Bookmark `json:"bookmarks"`
	}{
		Bookmarks: bookmarks,
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func handleAddBookmark(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Url string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.JSONErr(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := ctxhttp.Get(ctx, &crawler, input.Url)
	if err != nil {
		log.Error("cannot crawl",
			"url", input.Url,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body := make([]byte, 1024*20)
	if n, err := resp.Body.Read(body); err != nil && err != io.EOF {
		log.Error("cannot read crawler response",
			"url", input.Url,
			"error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	} else {
		body = body[:n]
	}

	title := pageTitle(body)

	var b Bookmark
	err = pg.DB(ctx).Get(&b, `
		INSERT INTO bookmarks (title, url, created)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
		RETURNING *
	`, title, input.Url, time.Now())
	if err != nil {
		log.Error("cannot create bookmark", "error", err.Error())
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	web.JSONResp(w, b, http.StatusCreated)
}

func pageTitle(b []byte) string {
	// <title></title> => 15 chars
	if node := findPageTitle(b); len(node) > 15 {
		title := string(node[7 : len(node)-8])
		return html.UnescapeString(title)
	}
	return ""
}

var findPageTitle = regexp.MustCompile(`<title>[^<]+</title>`).Find

var crawler = http.Client{
	Timeout: 15 * time.Second,
}
