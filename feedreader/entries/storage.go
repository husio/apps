package entries

import (
	"crypto/sha1"
	"encoding/base64"
	"time"

	"github.com/husio/x/storage/pg"
)

func ListEntries(s pg.Selector, limit, offset int) ([]*Entry, error) {
	var res []*Entry
	err := s.Select(&res, `
		SELECT * FROM entries
		ORDER BY created DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return res, pg.CastErr(err)
}

func AddEntry(ex pg.Execer, e *Entry) error {
	h := sha1.Sum([]byte(e.Content))
	e.ID = base64.StdEncoding.EncodeToString(h[:])
	if e.Created.IsZero() {
		e.Created = time.Now()
	}
	_, err := ex.Exec(`
		INSERT INTO entries (id, title, url, content, created)
		VALUES ($1, $2, $3, $4, $5)
	`, e.ID, e.Title, e.URL, e.Content, e.Created)
	return pg.CastErr(err)
}

type Entry struct {
	ID      string
	Title   string
	URL     string
	Content string
	Created time.Time
}
