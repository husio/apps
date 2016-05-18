package paste

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/husio/x/storage/pg"
	"github.com/husio/x/web"

	"golang.org/x/net/context"
)

func HandlePasteCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.StdJSONResp(w, http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		web.JSONErr(w, `"Content" is required"`, http.StatusBadRequest)
		return
	}

	db := pg.DB(ctx)
	paste, err := CreatePaste(db, Paste{Content: input.Content})
	if err != nil {
		log.Printf("cannot create paste: %s", err)
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	web.JSONResp(w, paste, http.StatusCreated)
}

func HandlePasteDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.ParseInt(web.Args(ctx).ByIndex(0), 10, 64)

	db := pg.DB(ctx)
	paste, err := PasteByID(db, pid)
	switch err {
	case nil:
		web.JSONResp(w, paste, http.StatusOK)
	case pg.ErrNotFound:
		web.StdJSONResp(w, http.StatusNotFound)
	default:
		log.Printf("cannot get paste %d: %s", pid, err)
		web.StdJSONResp(w, http.StatusInternalServerError)
	}
}

func HandlePasteList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := pg.DB(ctx)
	pastes, err := Pastes(db, 1000, 0)
	if err != nil {
		log.Printf("cannot list paste: %s", err)
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}
	resp := struct {
		Pastes []*Paste
	}{
		Pastes: pastes,
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func HandlePasteUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.StdJSONResp(w, http.StatusBadRequest)
		return
	}

	if input.Content == "" {
		web.JSONErr(w, `"Content" is required"`, http.StatusBadRequest)
		return
	}

	pid, _ := strconv.ParseInt(web.Args(ctx).ByIndex(0), 10, 64)

	db := pg.DB(ctx)
	paste, err := UpdatePaste(db, Paste{
		ID:      pid,
		Content: input.Content,
	})
	switch err {
	case nil:
		web.JSONResp(w, paste, http.StatusOK)
	case pg.ErrNotFound:
		web.StdJSONResp(w, http.StatusNotFound)
	default:
		log.Printf("cannot update paste %d: %s", pid, err)
		web.StdJSONResp(w, http.StatusInternalServerError)
	}
}

func HandlePasteDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	web.StdJSONResp(w, http.StatusNotImplemented)
}

func HandleRenderUI(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<meta http-equiv="x-ua-compatible" content="ie=edge">
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.2/css/bootstrap.min.css" integrity="sha384-y3tfxAZXuh4HwSYylfB+J125MxIs6mR5FOHamPBG064zB+AFeWH94NdvaCBm8qnd" crossorigin="anonymous">
		<style>
body, html { margin: 0; padding: 0; }
.entries {}
#pasteinput { width: 100%; height: 100%; }
		</style>
	  </head>
	  <body>
	  	<div class="container-fluid">
			<div class="col-md-2">
				<ul class="entries">
					<li>item 1</li>
					<li>item 2</li>
					<li>item 3</li>
					<li>item 4</li>
					<li>item 5</li>
				</ul>
			</div>
			<div class="col-md-10">
				<textarea id="pasteinput"></textarea>
			</div>
		</div>
	</body>

	<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.4/jquery.min.js"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.2/js/bootstrap.min.js" integrity="sha384-vZ2WRJMwsjRMW/8U7i6PWi6AlO1L79snBrmgiDpgIWJ82z8eA5lenwvxbMV1PAh7" crossorigin="anonymous"></script>
	<script src="/static/paste.js"></script>
</html>

	`)
}
