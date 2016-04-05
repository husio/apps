package auth

import (
	"time"

	"github.com/husio/x/storage/pg"
	"golang.org/x/crypto/bcrypt"
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

// SetPassword generate hash of given raw password representation and set it as
// account's PasswordHash attribute.
func (a *Account) SetPassword(raw []byte) error {
	ph, err := bcrypt.GenerateFromPassword(raw, bcrypt.DefaultCost+2)
	if err != nil {
		return err
	}
	a.PasswordHash = string(ph)
	return nil
}

// IsActive return true if account has not expired.
func (a *Account) IsActive() bool {
	return time.Now().Before(a.ValidTill)
}

// Accounts return list of all accounts, ordered by creation date.
func Accounts(s pg.Selector, limit, offset int64) ([]*Account, error) {
	var accs []*Account
	err := s.Select(&accs, `
		SELECT * FROM accounts
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	return accs, pg.CastErr(err)
}

// CreateAccount creates new account.
func CreateAccount(s pg.Selector, a Account) (*Account, error) {
	err := s.Select(&a, `
		INSERT INTO accounts (login, password_hash, scopes, created_at, valid_till)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, valid_till
	`, a.Login, a.PasswordHash, a.Scopes, time.Now(), a.ValidTill)
	if err != nil {
		return nil, pg.CastErr(err)
	}
	return &a, nil
}
