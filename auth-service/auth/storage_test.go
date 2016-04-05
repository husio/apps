package auth

import (
	"os"
	"reflect"
	"testing"

	"github.com/husio/x/storage/pg"
	"github.com/husio/x/storage/pg/pgtest"
	"github.com/jmoiron/sqlx"
)

// testDB create test database
func testDB(t *testing.T) (*sqlx.DB, func()) {
	db := pgtest.CreateDB(t, &pgtest.DBOpts{Host: os.Getenv("PG_HOST")})

	dbx := sqlx.NewDb(db, "postgres")
	pgtest.LoadSQL(t, dbx, "../schema.sql")
	pgtest.LoadSQL(t, dbx, "fixtures/storage_test_fixtures.sql")

	return dbx, func() { db.Close() }
}

func TestAccountByLogin(t *testing.T) {
	db, closedb := testDB(t)
	defer closedb()

	cases := map[string]struct {
		login string
		acc   *Account
		err   error
	}{
		"ok": {
			login: "bob@example.com",
			acc: &Account{
				ID:           1,
				Login:        "bob@example.com",
				PasswordHash: "xxx",
				Scopes:       []string{"admin"},
			},
		},
		"user_not_found": {
			login: "invalidlogin",
			err:   pg.ErrNotFound,
		},
	}

	for tname, tc := range cases {
		acc, err := AccountByLogin(db, tc.login)
		if err != tc.err {
			t.Errorf("%s: want %v, got %v", tname, tc.err, err)
			continue
		}

		if tc.acc == nil {
			if acc != nil {
				t.Errorf("%s: expect not account, got %v", tname, acc)
				continue
			}
		} else {
			if acc == nil {
				t.Errorf("%s: got no account, want %v", tname, tc.acc)
				continue
			}

			if tc.acc.ID != acc.ID ||
				tc.acc.Login != acc.Login ||
				tc.acc.PasswordHash != acc.PasswordHash ||
				!reflect.DeepEqual(tc.acc.Scopes, acc.Scopes) {
				t.Errorf("%s: want %+v, got %+v", tname, tc.acc, acc)
			}
		}
	}
}

func TestAccounts(t *testing.T) {
	db, closedb := testDB(t)
	defer closedb()

	cases := map[string]struct {
		limit  int64
		offset int64
		wantid []int64
	}{
		"ok_limit": {
			limit:  2,
			wantid: []int64{3, 2},
		},
		"ok_offset": {
			limit:  2,
			offset: 1,
			wantid: []int64{2, 4},
		},
		"no_result": {
			offset: 100,
		},
	}

	for tname, tc := range cases {
		accs, err := Accounts(db, tc.limit, tc.offset)
		if err != nil {
			t.Errorf("%s: cannot list accoutns: %s", tname, err)
			continue
		}

		if len(accs) != len(tc.wantid) {
			t.Errorf("%s: want %d accounts, got %d", tname, len(tc.wantid), len(accs))
			continue
		}
		var ids []int64
		for _, acc := range accs {
			ids = append(ids, acc.ID)
		}
		if !reflect.DeepEqual(tc.wantid, ids) {
			t.Errorf("%s: want ids %v, got %v", tname, tc.wantid, ids)
		}
	}
}
