package webcron

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
)

type Scheduler struct {
	storage storage
}

type storage interface {
	Add(Job) error
	List(limit, offset int) ([]Job, error)
	Del(string) error
}

func NewScheduler(s storage) *Scheduler {
	return &Scheduler{
		storage: s,
	}
}

// Add creates new job with unique ID assigned by the scheduler.
func (s *Scheduler) Add(job Job) (Job, error) {
	// TODO: adding job must reset Runner
	job.ID = genID()
	job.Created = time.Now()
	if err := s.storage.Add(job); err != nil {
		return job, fmt.Errorf("storage failure: %s", err)
	}
	return job, nil
}

// Del remove job with given ID. It returns ErrNotFound if job with given ID
// does not exist.
func (s *Scheduler) Del(jobID string) error {
	return s.storage.Del(jobID)
}

func (s *Scheduler) List(limit, offset int) ([]Job, error) {
	return s.storage.List(limit, offset)
}

func (s *Scheduler) Run(ctx context.Context) error {
	var next struct {
		time time.Time
		job  *Job
	}
	errc := make(chan error, 1)

	for {
		jobs, err := s.storage.List(1000, 0)
		if err != nil {
			return fmt.Errorf("cannot list jobs: %s", err)
		}
		if len(jobs) == 0 {
			time.Sleep(time.Second)
			continue
		}

		now := time.Now().Local()
		for _, job := range jobs {
			t := job.Schedule.Next(now)
			if next.time.IsZero() || t.Before(next.time) && !t.IsZero() {
				next.time = t
				next.job = &job
			}
		}

		select {
		case <-time.After(next.time.Sub(now)):
			go func(job *Job) {
				if err := job.Run(); err != nil {
					errc <- fmt.Errorf("job failed: %s", err)
				}
			}(next.job)
		case err := <-errc:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
