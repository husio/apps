package paste

import (
	"time"

	"github.com/husio/x/storage/pg"
)

type Paste struct {
	ID      int64
	Content string `json:",omitempty"`
	Created time.Time
	Updated time.Time
}

func CreatePaste(g pg.Getter, p Paste) (*Paste, error) {
	now := time.Now()
	err := g.Get(&p, `
		INSERT INTO pastes (content, created, updated)
		VALUES ($1, $2, $3)
		RETURNING *
	`, p.Content, now, now)
	return &p, pg.CastErr(err)
}

func Pastes(s pg.Selector, limit, offset int64) ([]*Paste, error) {
	var pastes []*Paste
	err := s.Select(&pastes, `
		SELECT
			id, created, updated,
			substring(content from 0 for 500) AS content
		FROM pastes
		ORDER BY updated DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return pastes, pg.CastErr(err)
}

func PasteByID(g pg.Getter, pid int64) (*Paste, error) {
	var p Paste
	err := g.Get(&p, `
		SELECT * FROM pastes
		WHERE id = $1
		LIMIT 1
	`, pid)
	return &p, pg.CastErr(err)
}

func UpdatePaste(g pg.Getter, p Paste) (*Paste, error) {
	now := time.Now()
	err := g.Get(&p, `
		UPDATE pastes
		SET content = $1, updated = $2
		WHERE id = $3
		RETURNING *
	`, p.Content, now, p.ID)
	return &p, pg.CastErr(err)
}
