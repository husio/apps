package auth

import (
	"time"

	"github.com/husio/x/storage/pg"
)

// AccountByLogin return account with given login.
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
	Login        string
	PasswordHash string `db:"password_hash" json:"-"`
	Scopes       pg.StringSlice
	ValidTill    time.Time `db:"valid_till"`
	CreatedAt    time.Time `db:"created_at"`
}

func (a *Account) IsActive() bool {
	return time.Now().Before(a.ValidTill)
}

func Accounts(s pg.Selector, limit, offset int64) ([]*Account, error) {
	var accs []*Account
	err := s.Select(&accs, `
		SELECT * FROM accounts
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return accs, pg.CastErr(err)
}
