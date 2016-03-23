package auth

import (
	"time"

	"github.com/husio/x/storage/pg"
)

func AccountByLogin(g pg.Getter, login string) (*Account, error) {
	var a Account
	err := g.Get(&a, `
		SELECT * FROM accounts
		WHERE login = $1
		LIMIT 1
	`, login)
	if err != nil {
		return nil, pg.CastErr(err)
	}
	return &a, nil
}

type Account struct {
	ID           int64
	Role         string
	Login        string
	PasswordHash string    `db:"password_hash" json:"-"`
	ValidTill    time.Time `db:"valid_till"`
	CreatedAt    time.Time `db:"created_at"`
}

func (a *Account) IsActive() bool {
	return time.Now().Before(a.ValidTill)
}
