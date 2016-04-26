package bb

import (
	"fmt"
	"net/http"
	"time"

	"github.com/husio/x/auth"
	"github.com/husio/x/log"
	"github.com/husio/x/web"
	"golang.org/x/net/context"
)

func NewApp(ctx context.Context) http.Handler {
	return &application{
		ctx: ctx,
		rt: web.NewRouter(web.Routes{
			{"GET", `/`, handleListTopics},
			{"GET", `/t`, handleListTopics},
			{"GET", `/t/new`, handleCreateTopic},
			{"POST", `/t/new`, handleCreateTopic},
			{"GET", `/t/{topic-id}`, handleTopicDetails},
			{"POST", `/t/{topic-id}/comment`, handleCreateComment},

			{"GET", `/login`, auth.LoginHandler("google")},
			{"GET", `/login/success`, auth.HandleLoginCallback},

			{web.AnyMethod, `.*`, handle404},
		}),
	}
}

type application struct {
	ctx context.Context
	rt  *web.Router
}

func handle404(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
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
