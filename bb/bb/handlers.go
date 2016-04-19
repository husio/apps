package bb

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/husio/x/log"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/web"

	"golang.org/x/net/context"
)

func handleListTopics(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	opts := TopicsOpts{
		Limit: 200,
		Tags:  r.URL.Query()["tag"],
	}

	db := pg.DB(ctx)
	topics, err := Topics(db, opts)
	if err != nil {
		log.Error("cannot list topics", "error", err.Error())
	}

	context := struct {
		Topics []*Topic
	}{
		Topics: topics,
	}
	render(w, "topic_list.tmpl", context)
}

func handleTopicDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	db := pg.DB(ctx)

	tid, _ := strconv.ParseInt(web.Args(ctx).ByIndex(0), 10, 64)
	topic, err := TopicByID(db, tid)
	switch err {
	case nil:
		// all good
	case pg.ErrNotFound:
		respond404(w, r)
		return
	default:
		log.Error("cannot get topic by ID",
			"topic", web.Args(ctx).ByIndex(0),
			"error", err.Error())
		respond500(w, r)
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	comments, err := Comments(db, CommentsOpts{
		Offset:  (page - 1) * 200,
		Limit:   200,
		TopicID: topic.TopicID,
	})
	if err != nil {
		log.Error("cannot get comments for topic",
			"topic", fmt.Sprint(topic.TopicID),
			"error", err.Error())
		respond500(w, r)
		return
	}

	context := struct {
		Topic    *Topic
		Comments []*Comment
	}{
		Topic:    topic,
		Comments: comments,
	}
	render(w, "topic_details.tmpl", context)
}

func handleCreateTopic(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	input := struct {
		Title   string
		Tags    []string
		Content string
	}{
		Title:   r.FormValue("title"),
		Tags:    strings.Fields(r.FormValue("tags")),
		Content: r.FormValue("content"),
	}

	var errs []string
	if r.Method == "POST" {
		if input.Title == "" {
			errs = append(errs, `"title" is required`)
		}
		if input.Content == "" {
			errs = append(errs, `"content" is required`)
		}
	}

	if r.Method == "GET" || len(errs) != 0 {
		render(w, "topic_create.tmpl", input)
		return
	}

	db := pg.DB(ctx)
	tx, err := db.Beginx()
	if err != nil {
		log.Error("cannot start transaction", "error", err.Error())
		respond500(w, r)
		return
	}
	defer tx.Rollback()

	topic := Topic{
		AuthorID: "author:1",
		Title:    input.Title,
		Tags:     input.Tags,
	}
	t, _, err := CreateTopicWithComment(tx, topic, input.Content)
	if err != nil {
		log.Error("cannot create topic with comment", "error", err.Error())
		respond500(w, r)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Error("cannot commit transaction", "error", err.Error())
		respond500(w, r)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/t/%d", t.TopicID), http.StatusSeeOther)
}

func handleCreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	input := struct {
		Content string
	}{
		Content: r.FormValue("content"),
	}

	var errs []string
	if len(input.Content) == 0 {
		errs = append(errs, `"content" is required`)
	}
	if len(errs) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%v", errs)
		return
	}

	db := pg.DB(ctx)
	tid, _ := strconv.ParseInt(web.Args(ctx).ByIndex(0), 10, 64)
	c, err := CreateComment(db, Comment{
		TopicID:  tid,
		Content:  input.Content,
		AuthorID: "user:1",
	})
	switch err {
	case nil:
		// ok
	default:
		log.Error("cannot create comment",
			"topic", fmt.Sprint(tid),
			"error", err.Error())
		respond500(w, r)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/t/%d", c.TopicID), http.StatusSeeOther)
}
