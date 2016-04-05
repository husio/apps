package auth

import (
	"net/http"

	"github.com/husio/x/web"
	"golang.org/x/net/context"
)

func NewApp(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{"POST", `/login`, HandleLogin},
			{"GET", `/keys/{key-id}`, HandlePublicKey},
			{"GET", `/accounts`, HandleListAccounts},
			{"POST", `/accounts`, HandleCreateAccount},
			{web.AnyMethod, `.*`, web.StdJSONHandler(http.StatusNotFound)},
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
