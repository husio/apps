package webcron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/robfig/cron"
)

type Job struct {
	ID          string           `json:"id"`
	Description string           `json:"description"`
	URL         string           `json:"url"`
	Schedule    Schedule         `json:"schedule"`
	Payload     *json.RawMessage `json:"payload"`
	Created     time.Time        `json:"created"`
}

func (j *Job) Run() error {
	var body io.Reader
	if j.Payload != nil {
		body = bytes.NewReader(*j.Payload)
	}
	resp, err := http.Post(j.URL, "application/json", body)
	if err != nil {
		return fmt.Errorf("request error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("invalid response %d: %s", resp.StatusCode, b)
	}
	return nil
}

// Start periodically run job, as defined in schedule attribute.
func (j *Job) Start(ctx context.Context) error {
	for {
		now := time.Now().Local()
		next := j.Schedule.Next(now)

		select {
		case <-time.After(next.Sub(now)):
			if err := j.Run(); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type Schedule struct {
	raw      string
	schedule cron.Schedule
}

func (s *Schedule) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	schedule, err := cron.Parse(raw)
	if err != nil {
		return err
	}
	s.raw = raw
	s.schedule = schedule
	return nil
}

func (s Schedule) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.raw)
}

func (s *Schedule) Next(now time.Time) time.Time {
	return s.schedule.Next(now)
}
