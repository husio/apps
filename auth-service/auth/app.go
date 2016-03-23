package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/husio/x/stamp"
	"github.com/husio/x/storage/pg"
	"github.com/husio/x/web"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

func NewApp(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{"POST", "/login", HandleLogin},
			{web.AnyMethod, ".*", web.StdJSONHandler(http.StatusNotFound)},
		}),
	}
}

type application struct {
	ctx context.Context
	rt  *web.Router
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.rt.ServeCtxHTTP(app.ctx, w, r)
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

	if !ValidPassword(input.Password, acc.PasswordHash) {
		web.StdJSONResp(w, http.StatusUnauthorized)
		return
	}

	if !acc.IsActive() {
		web.JSONErr(w, "account is not active", http.StatusUnauthorized)
		return
	}

	payload := struct {
		stamp.Claims
		ID       int64    `json:"id"`
		ClientIP string   `json:"ip"`
		Role     string   `json:"role"`
		Scopes   []string `json:"scop,omitempty"`
	}{
		Claims: stamp.Claims{
			ExpirationTime: min(
				time.Now().Add(2*time.Hour).Unix(),
				acc.ValidTill.Unix(),
			),
		},
		ID:       acc.ID,
		ClientIP: clientIP(r),
		Role:     acc.Role,
		Scopes:   nil,
	}

	signer := tokenSigner(ctx)
	token, err := stamp.Encode(signer, &payload)
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

// ValidPassword compare and return true if hash was generated for given password.
func ValidPassword(password, passHash string) bool {
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

// WithTokenSigner return context with signer attached.
func WithTokenSigner(ctx context.Context, s stamp.Signer) context.Context {
	return context.WithValue(ctx, "auth:signer", s)
}

// tokenSigner return signer attached to context. User WithTokenSigner prepare
// context.
func tokenSigner(ctx context.Context) stamp.Signer {
	s := ctx.Value("auth:signer")
	if s == nil {
		panic("token signer not present in context")
	}
	return s.(stamp.Signer)
}
