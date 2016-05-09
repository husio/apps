package gallery

import (
	"fmt"
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
	handleUI := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gallery/statics/index.html")
	}

	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{`/`, web.RedirectHandler("/ui/", http.StatusMovedPermanently), "GET"},
			{`/ui/.*`, handleUI, "GET"},

			{`/api/v1/images`, handleUploadImage, "PUT"},
			{`/api/v1/images`, handleListImages, "GET"},
			{`/api/v1/images/{id}\.jpg`, handleServeImage, "GET"},
			{`/api/v1/images/{id}/tags`, handleTagImage, "PUT"},
			{`/api/v1/images/{id}`, handleImageDetails, "GET"},

			{`.*`, handleStatics, "GET"},

			{`.*`, web.StdJSONHandler(http.StatusNotFound), web.AnyMethod},
		}),
	}
}

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rl := responseLogger{
		start: time.Now(),
		code:  http.StatusOK,
		w:     w,
	}
	app.rt.ServeCtxHTTP(app.ctx, &rl, r)
	work := time.Now().Sub(rl.start)

	writelog := log.Debug
	if rl.code >= 500 {
		writelog = log.Error
	}
	writelog("request served",
		"workTime", work.String(),
		"code", fmt.Sprint(rl.code),
		"method", r.Method,
		"url", r.URL.String())
}

type responseLogger struct {
	start time.Time
	code  int
	w     http.ResponseWriter
}

func (rl *responseLogger) Header() http.Header {
	return rl.w.Header()
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	return rl.w.Write(b)
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.code = code
	rl.w.WriteHeader(code)
}
