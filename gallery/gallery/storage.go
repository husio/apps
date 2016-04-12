package gallery

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/husio/x/storage/qb"
	"github.com/husio/x/storage/sq"
)

type Image struct {
	ImageID string    `db:"image_id" json:"imageId"`
	Width   int       `db:"width"    json:"width"`
	Height  int       `db:"height"   json:"height"`
	Created time.Time `db:"created"  json:"created"`
}

func Images(s sq.Selector, opts ImagesOpts) ([]*Image, error) {
	var q qb.Query
	if len(opts.Tags) == 0 {
		q = qb.Q("SELECT * FROM images i")
	} else {
		q = qb.Q("SELECT * FROM images i INNER JOIN tags t")
		for _, kv := range opts.Tags {
			q.Where("t.name = ? AND t.value = ?", kv.Key, kv.Value)
		}
	}

	q.OrderBy("i.created DESC").Limit(opts.Limit, opts.Offset)
	query, args := q.Build()

	var imgs []*Image
	err := s.Select(&imgs, query, args...)
	return imgs, sq.CastErr(err)
}

type ImagesOpts struct {
	Limit  int64
	Offset int64
	Tags   []KeyValue
}

type KeyValue struct {
	Key   string
	Value string
}

func CreateImage(e sq.Execer, img Image) (*Image, error) {
	if img.Created.IsZero() {
		img.Created = time.Now()
	}

	_, err := e.Exec(`
		INSERT INTO images (image_id, width, height, created)
		VALUES (?, ?, ?, ?)
	`, img.ImageID, img.Width, img.Height, img.Created)
	return &img, sq.CastErr(err)
}

func ImageByID(g sq.Getter, imageID string) (*Image, error) {
	var img Image
	err := g.Get(&img, `
		SELECT * FROM images
		WHERE image_id = ?
		LIMIT 1
	`, imageID)
	if err != nil {
		return nil, sq.CastErr(err)
	}
	return &img, nil
}

func ImageExists(g sq.Getter, imageID string) error {
	var ok int
	err := g.Get(&ok, `
		SELECT 1 FROM images
		WHERE image_id = ?
		LIMIT 1
	`, imageID)
	return sq.CastErr(err)
}

type Tag struct {
	TagID   string    `db:"tag_id"   json:"tagId"`
	ImageID string    `db:"image_id" json:"imageId"`
	Name    string    `db:"name"     json:"name"`
	Value   string    `db:"value"    json:"value"`
	Created time.Time `db:"created"  json:"created"`
}

func CreateTag(e sq.Execer, tag Tag) (*Tag, error) {
	oid := sha256.New()
	fmt.Fprint(oid, tag.ImageID)
	fmt.Fprint(oid, tag.Name)
	fmt.Fprint(oid, tag.Value)
	tag.TagID = encode(oid)

	if tag.Created.IsZero() {
		tag.Created = time.Now()
	}

	_, err := e.Exec(`
		INSERT INTO tags (tag_id, image_id, name, value, created)
		VALUES (?, ?, ?, ?, ?)
	`, tag.TagID, tag.ImageID, tag.Name, tag.Value, tag.Created)
	return &tag, sq.CastErr(err)
}

func encode(h hasher) string {
	s := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return strings.TrimRight(s, "=")
}

type hasher interface {
	Sum(b []byte) []byte
}
