package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/husio/x/stamp"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/storage/pgtest"
	"golang.org/x/net/context"
)

func TestHandleGetDocument(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		db   *pgtest.DB
		body string
		code int
	}{
		"ok": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack: []pgtest.ResultMock{
					{"Get", &Account{
						ID:           1,
						Role:         "admin",
						Login:        "bob@example.com",
						PasswordHash: "hash:secret",
						ValidTill:    now.Add(time.Hour),
						CreatedAt:    now.Add(-time.Hour),
					}, nil},
				},
			},
			body: `{"login": "bob@example.com", "password": "secret"}`,
			code: http.StatusOK,
		},
		"no_login": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack:  []pgtest.ResultMock{},
			},
			body: `{"login": "", "password": "secret"}`,
			code: http.StatusBadRequest,
		},
		"no_password": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack:  []pgtest.ResultMock{},
			},
			body: `{"login": "bob@example.com", "password": ""}`,
			code: http.StatusBadRequest,
		},
		"expired_account": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack: []pgtest.ResultMock{
					{"Get", &Account{
						ID:           1,
						Role:         "admin",
						Login:        "bob@example.com",
						PasswordHash: "hash:secret",
						ValidTill:    now.Add(-666 * time.Minute),
						CreatedAt:    now.Add(-time.Hour),
					}, nil},
				},
			},
			body: `{"login": "bob@example.com", "password": "secret"}`,
			code: http.StatusUnauthorized,
		},
		"invalid_login": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack: []pgtest.ResultMock{
					{"Get", nil, pg.ErrNotFound},
				},
			},
			body: `{"login": "bob@example.com", "password": "secret"}`,
			code: http.StatusUnauthorized,
		},
		"invalid_password": {
			db: &pgtest.DB{
				Fatalf: t.Fatalf,
				Stack: []pgtest.ResultMock{
					{"Get", &Account{
						ID:           1,
						Role:         "admin",
						Login:        "bob@example.com",
						PasswordHash: "xxxx",
						ValidTill:    now.Add(-666 * time.Minute),
						CreatedAt:    now.Add(-time.Hour),
					}, nil},
				},
			},
			body: `{"login": "bob@example.com", "password": "secret"}`,
			code: http.StatusUnauthorized,
		},
	}

	defer useFakePasswordHashComparator()()

	for tname, tc := range cases {
		ctx := context.Background()
		ctx = pgtest.WithDB(ctx, tc.db)
		ctx = WithTokenSigner(ctx, &xSigner{nil})
		app := NewApp(ctx)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/login", strings.NewReader(tc.body))
		app.ServeHTTP(w, r)

		if w.Code != tc.code {
			t.Errorf("%s: want %d, got %d: %s", tname, tc.code, w.Code, w.Body)
		}
	}
}

func TestLoginToken(t *testing.T) {
	now := time.Now()

	defer useFakePasswordHashComparator()()

	ctx := context.Background()
	ctx = pgtest.WithDB(ctx, &pgtest.DB{
		Fatalf: t.Fatalf,
		Stack: []pgtest.ResultMock{
			{"Get", &Account{
				ID:           519,
				Role:         "admin",
				Login:        "bob@example.com",
				PasswordHash: "hash:secret",
				ValidTill:    now.Add(time.Hour),
				CreatedAt:    now.Add(-time.Hour),
			}, nil},
		},
	})
	ctx = WithTokenSigner(ctx, &xSigner{nil})
	app := NewApp(ctx)

	w := httptest.NewRecorder()
	body := strings.NewReader(`
		{
			"login": "bob@example.com",
			"password": "secret"
		}
	`)
	r, _ := http.NewRequest("POST", "/login", body)
	r.Header.Set("X-Forwarded-For", "6.6.6.6")
	app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", w.Code, w.Body)
	}

	var resp struct {
		Token string
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("cannot decode response: %s: %s", w.Body, err)
	}

	var payload struct {
		ID       int64  `json:"id"`
		ClientIP string `json:"ip"`
		Role     string `json:"role"`
	}
	if err := stamp.Decode(&xSigner{nil}, &payload, []byte(resp.Token)); err != nil {
		t.Fatalf("cannot decode token %q: %s", resp.Token, err)
	}

	if payload.Role != "admin" || payload.ID != 519 || payload.ClientIP != "6.6.6.6" {
		t.Fatalf("invalid payload: %+v", payload)
	}
}

func useFakePasswordHashComparator() func() {
	// use mock password hash comparator
	cmp := compareAndHashPassword
	compareAndHashPassword = func(p, h []byte) error {
		if string(p) != "hash:"+string(h) {
			return errors.New("invalid")
		}
		return nil
	}
	return func() {
		compareAndHashPassword = cmp
	}
}

// xSigner implements stamp.Signer interface, providing easy to predict
// signature algorithm
type xSigner struct {
	err error
}

func (xSigner) Algorithm() string {
	return "x"
}
func (x *xSigner) Sign(data []byte) ([]byte, error) {
	return []byte("x"), x.err
}

func (x *xSigner) Verify(signature, data []byte) error {
	if string(signature) != "x" {
		return stamp.ErrInvalidSignature
	}
	return x.err
}
