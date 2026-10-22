package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"queuelet/internal/client"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	ID      int64          `json:"id"`
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload"`
	Attempt int            `json:"attempt"`
	Token   string         `json:"lease_token"`
}

func Run(ctx context.Context, c *client.Client, name, queue string, kinds []string) error {
	var reg struct {
		ID int64 `json:"worker_id"`
	}
	if e := c.JSON("POST", "/v1/workers", map[string]any{"name": name, "kinds": kinds}, &reg); e != nil {
		return e
	}
	log.Printf("worker %s registered", name)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		var batch struct {
			Jobs []Job `json:"jobs"`
		}
		if e := c.JSON("POST", "/v1/queues/"+queue+"/lease", map[string]any{"worker_id": reg.ID, "max_jobs": 1}, &batch); e != nil {
			log.Print(e)
			time.Sleep(time.Second)
			continue
		}
		for _, j := range batch.Jobs {
			slog.Info("processing", "event", "processing", "job_id", j.ID, "queue", queue, "worker", reg.ID, "attempt", j.Attempt, "kind", j.Kind)
			workCtx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})
			heartbeatDone := make(chan struct{})
			go func(j Job) {
				defer close(heartbeatDone)
				t := time.NewTicker(time.Second)
				defer t.Stop()
				for {
					select {
					case <-done:
						return
					case <-t.C:
						b, status, err := c.Do("POST", fmt.Sprintf("/v1/jobs/%d/heartbeat", j.ID), map[string]any{"lease_token": j.Token, "extend_seconds": 30})
						var observation struct {
							CancelRequested bool `json:"cancel_requested"`
						}
						if err == nil && status == http.StatusConflict && json.Unmarshal(b, &observation) == nil && observation.CancelRequested {
							slog.Info("cancellation observed", "event", "cancel_requested", "job_id", j.ID)
							cancel()
							return
						}
					}
				}
			}(j)
			result, e := perform(workCtx, j)
			close(done)
			cancel()
			<-heartbeatDone
			verb := "ack"
			body := map[string]any{"lease_token": j.Token, "result": result}
			if e != nil {
				verb = "nack"
				body = map[string]any{"lease_token": j.Token, "error": e.Error(), "retry": true}
			}
			_, status, e := c.Do("POST", fmt.Sprintf("/v1/jobs/%d/%s", j.ID, verb), body)
			slog.Info("completed", "event", verb, "job_id", j.ID, "queue", queue, "worker", reg.ID, "attempt", j.Attempt, "status", status, "error", e)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
func perform(ctx context.Context, j Job) (any, error) {
	switch {
	case j.Kind == "echo":
		return j.Payload, nil
	case j.Kind == "sleep":
		number, _ := j.Payload["seconds"].(json.Number)
		seconds, _ := number.Float64()
		timer := time.NewTimer(time.Duration(seconds * float64(time.Second)))
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
		return j.Payload, nil
	case j.Kind == "crash":
		os.Exit(3)
	case strings.HasPrefix(j.Kind, "fail"):
		number, _ := j.Payload["n"].(json.Number)
		n, _ := number.Float64()
		if n == 0 {
			v, _ := strconv.Atoi(strings.TrimPrefix(j.Kind, "fail"))
			n = float64(v)
		}
		if j.Attempt <= int(n) {
			return nil, fmt.Errorf("requested failure %d", j.Attempt)
		}
		return j.Payload, nil
	case j.Kind == "http":
		url, _ := j.Payload["url"].(string)
		req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
		if e != nil {
			return nil, e
		}
		resp, e := (&http.Client{Timeout: 5 * time.Second}).Do(req)
		if e != nil {
			return nil, e
		}
		defer resp.Body.Close()
		b, e := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if e != nil {
			return nil, e
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("remote status %d", resp.StatusCode)
		}
		var v any
		if json.Unmarshal(b, &v) != nil {
			v = string(b)
		}
		return v, nil
	}
	return nil, fmt.Errorf("unknown kind %s", j.Kind)
}
