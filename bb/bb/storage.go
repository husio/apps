package bb

import (
	"time"

	"github.com/husio/x/auth"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/storage/qb"
	"github.com/jmoiron/sqlx"
)

type Topic struct {
	TopicID  int64          `db:"topic_id"  json:"topic_id"`
	AuthorID int64          `db:"author_id" json:"author_id"`
	Title    string         `db:"title"     json:"title"`
	Tags     pg.StringSlice `db:"tags"      json:"tags"`
	Created  time.Time      `db:"created"   json:"created"`
	Updated  time.Time      `db:"updated"   json:"updated"`
}

type TopicWithAuthor struct {
	*Topic
	*auth.Account
}

// Topics return slice of topics that match given query criteria.
func Topics(s pg.Selector, o TopicsOpts) ([]*TopicWithAuthor, error) {
	if o.Offset < 0 {
		o.Offset = 0
	}
	q := qb.Q(`
		SELECT t.*, a.*
		FROM topics t LEFT JOIN accounts a ON t.author_id = a.account_id
	`).Limit(o.Limit, o.Offset).OrderBy("t.created ASC")

	if len(o.Tags) > 0 {
		// TODO
	}
	if !o.OlderThan.IsZero() {
		q.Where("t.created < ?", o.OlderThan)
	}

	query, args := q.Build()
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var topics []*TopicWithAuthor
	if err := s.Select(&topics, query, args...); err != nil {
		return nil, pg.CastErr(err)
	}
	return topics, nil
}

type TopicsOpts struct {
	Limit     int64
	Offset    int64
	Tags      []string
	OlderThan time.Time
}

func TopicByID(g pg.Getter, topicID int64) (*Topic, error) {
	var t Topic
	err := g.Get(&t, `
		SELECT * FROM topics WHERE topic_id = $1 LIMIT 1
	`, topicID)
	if err != nil {
		return &t, pg.CastErr(err)
	}
	return &t, nil
}

func CreateTopic(g pg.Getter, t Topic) (*Topic, error) {
	now := time.Now()

	if t.Created.IsZero() {
		t.Created = now
	}
	t.Updated = now

	var tid int64
	err := g.Get(&tid, `
		INSERT INTO topics (author_id, title, tags, created, updated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING topic_id
	`, t.AuthorID, t.Title, t.Tags, t.Created, t.Updated)
	if err != nil {
		return nil, pg.CastErr(err)
	}
	t.TopicID = tid
	return &t, nil
}

type Comment struct {
	CommentID int64     `db:"comment_id" json:"comment_id"`
	TopicID   int64     `db:"topic_id"   json:"topic_id"`
	AuthorID  int64     `db:"author_id"  json:"author_id"`
	Content   string    `db:"content"    json:"content"`
	Created   time.Time `db:"created"    json:"created"`
	Updated   time.Time `db:"updated"    json:"updated"`
}

func CreateComment(g pg.Getter, c Comment) (*Comment, error) {
	now := time.Now()

	if c.Created.IsZero() {
		c.Created = now
	}
	c.Updated = now

	var cid int64
	err := g.Get(&cid, `
		INSERT INTO comments (topic_id, author_id, content, created, updated)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING comment_id
	`, c.TopicID, c.AuthorID, c.Content, c.Created, c.Updated)
	if err != nil {
		return nil, pg.CastErr(err)
	}
	c.CommentID = cid
	return &c, nil
}

func CreateTopicWithComment(g pg.Getter, topic Topic, comment string) (*Topic, *Comment, error) {
	t, err := CreateTopic(g, topic)
	if err != nil {
		return nil, nil, err
	}
	c, err := CreateComment(g, Comment{
		TopicID:  t.TopicID,
		AuthorID: t.AuthorID,
		Content:  comment,
		Created:  t.Created,
		Updated:  t.Updated,
	})
	if err != nil {
		return nil, nil, err
	}
	return t, c, nil
}

// Comments return slice of comments that match given query criteria.
func Comments(s pg.Selector, o CommentsOpts) ([]*Comment, error) {
	if o.Offset < 0 {
		o.Offset = 0
	}
	q := qb.Q("SELECT * FROM comments").Limit(o.Limit, o.Offset).OrderBy("created DESC")
	if o.TopicID != 0 {
		q.Where("topic_id = ?", o.TopicID)
	}
	if o.AuthorID != "" {
		q.Where("author_id = ?", o.AuthorID)
	}
	query, args := q.Build()
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var comments []*Comment
	if err := s.Select(&comments, query, args...); err != nil {
		return nil, pg.CastErr(err)
	}

	return comments, nil
}

type CommentsOpts struct {
	Limit    int64
	Offset   int64
	TopicID  int64
	AuthorID string
}
