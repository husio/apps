package qux

import (
	"net/http"

	"github.com/husio/x/web"
	"golang.org/x/net/context"
)

type application struct {
	ctx context.Context
	rt  *web.Router
}

func NewApplication(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{"POST", `/v1/queues/{queue}`, handleSubscribe},
			{"DELETE", `/v1/queues/{queue}`, handleUnsubscribe},
			{"POST", `/v1/ping`, handlePong},
		}),
	}
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.rt.ServeCtxHTTP(app.ctx, w, r)
}
