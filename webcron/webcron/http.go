package webcron

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/husio/x/web"

	"golang.org/x/net/context"
)

type httpUI struct {
	scheduler scheduler
}

type scheduler interface {
	Add(Job) (Job, error)
	Del(string) error
	List(limit, offset int) ([]Job, error)
}

var router = web.NewRouter("", web.Routes{
	web.GET(`/jobs`, "job-list", handleListJobs),
	web.POST(`/jobs`, "job-create", handleCreateJob),
	web.DELETE(`/jobs/{job-id}`, "job-delete", handleDeleteJob),
	web.ANY(`.*`, "", handleNotFound),
})

func NewHandler(scheduler scheduler) http.Handler {
	return &httpUI{
		scheduler: scheduler,
	}
}

func (ui *httpUI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ui.ServeCtxHTTP(context.Background(), w, r)
}

func (ui *httpUI) ServeCtxHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, "scheduler", ui.scheduler)
	router.ServeCtxHTTP(ctx, w, r)
}

// ctxScheduler return scheduler present in given context. It panics if
// scheduler is missing.
func ctxScheduler(ctx context.Context) scheduler {
	s := ctx.Value("scheduler")
	if s == nil {
		panic("scheduler not present in the context")
	}
	return s.(scheduler)
}

func handleNotFound(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// trailing slash is not allowed
	if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
		path := strings.TrimSuffix(r.URL.Path, "/")
		web.JSONRedirect(w, path, http.StatusMovedPermanently)
		return
	}

	web.StdJSONErr(w, http.StatusNotFound)
}

func handleListJobs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		limit  int = 50
		offset int = 0
		errs   []string
	)

	errs = append(errs, parseInt(&limit, r.URL.Query(), "limit")...)
	errs = append(errs, parseInt(&offset, r.URL.Query(), "offset")...)
	if len(errs) != 0 {
		web.JSONErrs(w, errs, http.StatusBadRequest)
		return
	}

	scheduler := ctxScheduler(ctx)
	if jobs, err := scheduler.List(limit, offset); err != nil {
		web.JSONErr(w, err.Error(), http.StatusBadRequest)
	} else {
		resp := struct {
			Jobs []Job `json:"jobs"`
		}{
			Jobs: jobs,
		}
		web.JSONResp(w, resp, http.StatusOK)
	}
}

func parseInt(dest *int, getter getter, name string) []string {
	raw := getter.Get(name)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return []string{fmt.Sprintf("%q value must be integer: %s", err.Error())}
	}

	if n < 0 {
		return []string{fmt.Sprintf("%q value must greater or equal zero", name)}
	}

	*dest = int(n)
	return nil
}

type getter interface {
	Get(string) string
}

func handleCreateJob(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		web.JSONErr(w, err.Error(), http.StatusBadRequest)
		return
	}

	if job.URL == "" {
		web.JSONErr(w, "'url' is required", http.StatusBadRequest)
		return
	}

	scheduler := ctxScheduler(ctx)
	if job, err := scheduler.Add(job); err != nil {
		web.JSONErr(w, err.Error(), http.StatusInternalServerError)
	} else {
		web.JSONResp(w, job, http.StatusOK)
	}
}

func handleDeleteJob(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	scheduler := ctxScheduler(ctx)
	jobID := web.Args(ctx).ByName("job-id")

	switch err := scheduler.Del(jobID); err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	case ErrNotFound:
		web.StdJSONErr(w, http.StatusNotFound)
	default:
		web.JSONErr(w, err.Error(), http.StatusInternalServerError)
	}
}
