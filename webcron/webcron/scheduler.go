package webcron

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
)

type Scheduler struct {
	ctx     context.Context
	stop    func()
	errc    chan error
	storage storage
}

type storage interface {
	Add(Job) error
	List(limit, offset int) ([]Job, error)
	Del(string) error
}

func RunScheduler(ctx context.Context, storage storage) (*Scheduler, error) {
	ctx, stop := context.WithCancel(ctx)
	s := &Scheduler{
		ctx:     ctx,
		storage: storage,
		errc:    make(chan error, 32),
		stop:    stop,
	}

	jobs, err := s.storage.List(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot list jobs: %s", err)
	}
	for _, job := range jobs {
		go runJob(ctx, job, s.errc)
	}

	return s, nil
}

func (s *Scheduler) Stop() {
	s.stop()
}

func (s *Scheduler) Errc() <-chan error {
	return s.errc
}

// Add creates new job with unique ID assigned by the scheduler.
func (s *Scheduler) Add(job Job) (Job, error) {
	job.ID = genID()
	job.Created = time.Now()
	if err := s.storage.Add(job); err != nil {
		return job, fmt.Errorf("storage failure: %s", err)
	}
	go runJob(s.ctx, job, s.errc)
	return job, nil
}

// Del remove job with given ID. It returns ErrNotFound if job with given ID
// does not exist.
func (s *Scheduler) Del(jobID string) error {
	// TODO: stop scheduled job
	return s.storage.Del(jobID)
}

func (s *Scheduler) List(limit, offset int) ([]Job, error) {
	return s.storage.List(limit, offset)
}

func runJob(ctx context.Context, job Job, errc chan<- error) {
	for {
		if err := job.Start(ctx); err != nil {
			select {
			case errc <- err:
			case <-ctx.Done():
				return
			default:
				log.Print("ignoring error: cannot push through channel")
				// ignore errors that cannot be pushed through channel
			}
		}
	}
}
