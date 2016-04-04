package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/husio/x/stamp"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/web"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

func HandlePublicKey(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	kid := web.Args(ctx).ByIndex(0)
	key, ok := keyManager(ctx).KeyByID(kid)
	if !ok {
		web.JSONErr(w, "key not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/pkix-crl")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pub"`, kid))
	fmt.Fprint(w, key)
}

// HandleLogin authenticate user using login/password and if successful, return
// autorization token.
// Credentials can be passed by using either basic auth or JSON encoded body.
func HandleLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string
		Password string
	}

	if login, pass, ok := r.BasicAuth(); ok {
		input.Login = login
		input.Password = pass
	} else {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			web.JSONErr(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var errs []string
	if input.Login == "" {
		errs = append(errs, `"login" is required`)
	}
	if len(input.Login) > 100 {
		errs = append(errs, `"login" too long`)
	}
	if input.Password == "" {
		errs = append(errs, `"password" is required`)
	}
	if len(input.Password) > 80 {
		errs = append(errs, `"password" too long`)
	}
	if len(errs) != 0 {
		web.JSONErrs(w, errs, http.StatusBadRequest)
		return
	}

	db := pg.DB(ctx)
	acc, err := AccountByLogin(db, input.Login)
	switch err {
	case nil:
		// all good
	case pg.ErrNotFound:
		// we don't want to allow to probe user logins
		web.StdJSONResp(w, http.StatusUnauthorized)
		return
	default:
		log.Printf("cannot get user %q by login: %s", input.Login, err)
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	if !validPassword(input.Password, acc.PasswordHash) {
		web.StdJSONResp(w, http.StatusUnauthorized)
		return
	}

	if !acc.IsActive() {
		web.JSONErr(w, "account is not active", http.StatusUnauthorized)
		return
	}

	payload := struct {
		stamp.Claims
		ID       int64    `json:"userid"`
		ClientIP string   `json:"ip,omitempty"`
		Scopes   []string `json:"scopes,omitempty"`
	}{
		Claims: stamp.Claims{
			ExpirationTime: min(
				time.Now().Add(2*time.Hour).Unix(),
				acc.ValidTill.Unix(),
			),
		},
		ID:       acc.ID,
		ClientIP: clientIP(r),
		Scopes:   acc.Scopes,
	}

	token, err := keyManager(ctx).Vault().Encode(&payload)
	if err != nil {
		log.Printf("cannot encode token: %s", err)
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	resp := struct {
		Token string `json:"token"`
	}{
		Token: string(token),
	}
	web.JSONResp(w, resp, http.StatusOK)
}

// validPassword compare and return true if hash was generated for given password.
func validPassword(password, passHash string) bool {
	return compareAndHashPassword([]byte(passHash), []byte(password)) == nil
}

// alias so that tests can mock it
var compareAndHashPassword = bcrypt.CompareHashAndPassword

// clientIP return request client's IP with priority for information from the
// header.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func min(first int64, rest ...int64) int64 {
	min := first
	for _, n := range rest {
		if n < min {
			min = n
		}
	}
	return min
}

func WithKeyManager(ctx context.Context, km *KeyManager) context.Context {
	return context.WithValue(ctx, "auth:keymanager", km)
}

func keyManager(ctx context.Context) *KeyManager {
	km := ctx.Value("auth:keymanager")
	if km == nil {
		panic("key manager not present in context")
	}
	return km.(*KeyManager)
}

func HandleListAccounts(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if !isAdmin(keyManager(ctx).Vault(), r) {
		web.StdJSONResp(w, http.StatusUnauthorized)
		return
	}

	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	db := pg.DB(ctx)
	accs, err := Accounts(db, 200, offset)
	if err != nil {
		log.Printf("cannot list accounts: %s", err)
		web.StdJSONResp(w, http.StatusInternalServerError)
		return
	}

	resp := struct {
		Accounts []*Account `json:"accounts"`
	}{
		Accounts: accs,
	}
	web.JSONResp(w, resp, http.StatusOK)
}

func isAdmin(v *stamp.Vault, r *http.Request) bool {
	var payload struct {
		Scopes []string `json:"scopes"`
	}
	if err := authPayload(v, r, &payload); err != nil {
		return false
	}
	for _, s := range payload.Scopes {
		if s == "admin" {
			return true
		}
	}
	return false
}

func authPayload(v *stamp.Vault, r *http.Request, payload interface{}) error {
	token := r.URL.Query().Get("authToken")
	if token == "" {
		if fs := strings.Fields(r.Header.Get("Authorization")); len(fs) == 2 {
			token = fs[1]
		}
	}
	if token == "" {
		return errors.New("no token")
	}
	return v.Decode(&payload, []byte(token))
}
