package storage

import (
	"time"

	"github.com/husio/x/storage/pg"
)

type Source struct {
	SourceID int64     `db:"source_id"`
	Link     string    `db:"link"`
	Created  time.Time `db:"created"`
}

type Subscription struct {
	SubscriptionID int64     `db:"subscription_id"`
	SubscriberID   int64     `db:"subscriber_id"`
	SourceID       int64     `db:"source_id" json:"-"`
	Created        time.Time `db:"created"`

	Source Source
}

type Entry struct {
	EntryID  int64     `db:"entry_id"`
	SourceID int64     `db:"source_id"`
	Link     string    `db:"link"`
	Title    string    `db:"title"`
	Created  time.Time `db:"created"`

	Source Source
}

func Subscribe(g pg.Getter, subscriber int64, link string) (*Subscription, error) {
	now := time.Now()

	var sub Subscription
	err := g.Get(&sub.Source, `
		INSERT INTO sources (link, created)
			VALUES ($1, $2)
		ON CONFLICT DO NOTHING
		RETURNING *
	`, link, now)
	if err != nil {
		return nil, pg.CastErr(err)
	}

	err = g.Get(&sub, `
		INSERT INTO subscriptions (subscriber_id, source_id, created)
			VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
		RETURNING *
	`, subscriber, sub.Source.SourceID, now)
	if err != nil {
		return nil, pg.CastErr(err)
	}

	return &sub, err
}

func Subscriptions(s pg.Selector, subscriber int64, limit, offset int64) ([]*Subscription, error) {
	var subs []*Subscription
	err := s.Select(&subs, `
		SELECT sub.*, src.*
		FROM subscriptions sub
			INNER JOIN sources src ON sub.source_id = src.source_id
		WHERE
			sub.subscriber_id = $1
		ORDER BY sub.created DESC
		LIMIT $2 OFFSET $3
	`, subscriber, limit, offset)
	return subs, pg.CastErr(err)
}

func Entries(s pg.Selector, subscriber int64, limit, offset int64) ([]*Entry, error) {
	var entries []*Entry
	err := s.Select(&entries, `
		SELECT e.*, src.*
		FROM entries e
			OUTER JOIN sources src ON src.source_id = e.source_id
			OUTER JOIN subscriptions sub ON sub.source_id = src.source_id
		WHERE
			sub.subscriber_id = $1
		ORDER BY e.created DESC
		LIMIT $2 OFFSET $3
	`, subscriber, limit, offset)
	return entries, pg.CastErr(err)
}
