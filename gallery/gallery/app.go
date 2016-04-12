package gallery

import (
	"net/http"
	"time"

	"github.com/husio/x/log"
	"github.com/husio/x/web"
	"golang.org/x/net/context"
)

type application struct {
	ctx context.Context
	rt  *web.Router
}

func NewApplication(ctx context.Context) http.Handler {
	statics := http.FileServer(http.Dir("."))
	handleStatics := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		statics.ServeHTTP(w, r)
	}

	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{

			{"PUT", `/api/v1/images`, handleUploadImage},
			{"GET", `/api/v1/images`, handleListImages},
			{"GET", `/api/v1/images/{id}\.jpg`, handleServeImage},
			{"GET", `/api/v1/images/{id}/tags`, handleImageTags},
			{"PUT", `/api/v1/images/{id}/tags`, handleTagImage},

			{"GET", `.*`, handleStatics},

			{web.AnyMethod, `.*`, web.StdJSONHandler(http.StatusNotFound)},
		}),
	}
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	app.rt.ServeCtxHTTP(app.ctx, w, r)
	work := time.Now().Sub(start)
	log.Debug("request served",
		"workTime", work.String(),
		"method", r.Method,
		"url", r.URL.String())
}
