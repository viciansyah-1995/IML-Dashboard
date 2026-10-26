package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iml/internal/auth"
	"iml/internal/config"
	"iml/internal/extraction"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Worker struct {
	Base, Token, ID string
	HTTP            *http.Client
}
type Claim struct {
	Resource     string                     `json:"resource"`
	MetricKeys   []string                   `json:"metric_keys"`
	Operator     extraction.OperatorContext `json:"operator"`
	JobID        string                     `json:"job_id"`
	Mode         string                     `json:"mode"`
	Scenario     string                     `json:"scenario"`
	Attempt      int                        `json:"attempt"`
	Lease        string                     `json:"lease_token"`
	Channel      string                     `json:"channel"`
	Scope        string                     `json:"scope"`
	Attachments  []extraction.Attachment    `json:"attachments"`
	LeaseSeconds int                        `json:"lease_seconds"`
}

func (w Worker) request(ctx context.Context, method, path, lease string, payload any) ([]byte, int, error) {
	var reader io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		reader = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(ctx, method, w.Base+path, reader)
	if e != nil {
		return nil, 0, e
	}
	req.Header.Set("Authorization", "Bearer "+w.Token)
	req.Header.Set("Content-Type", "application/json")
	if lease != "" {
		req.Header.Set("X-Lease-Token", lease)
	}
	res, e := w.HTTP.Do(req)
	if e != nil {
		return nil, 0, e
	}
	defer res.Body.Close()
	b, e := io.ReadAll(io.LimitReader(res.Body, 11<<20))
	if res.StatusCode >= 400 {
		var problem struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		json.Unmarshal(b, &problem)
		return b, res.StatusCode, fmt.Errorf("internal API HTTP %d: %s %s", res.StatusCode, problem.Code, problem.Message)
	}
	return b, res.StatusCode, e
}
func main() {
	config.LoadEnv(".local/worker.env")
	token := os.Getenv("WORKER_TOKEN")
	if token == "" {
		slog.Error("WORKER_TOKEN missing; make worker-env")
		os.Exit(1)
	}
	base := os.Getenv("WORKER_API_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}
	timeout := 90 * time.Second
	if n, e := time.ParseDuration(os.Getenv("OPENCLAW_TIMEOUT_SECONDS") + "s"); e == nil {
		timeout = n
	}
	w := Worker{Base: base, Token: token, ID: auth.ID(), HTTP: &http.Client{Timeout: timeout + 10*time.Second}}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	concurrency := 1
	fmt.Sscanf(os.Getenv("WORKER_CONCURRENCY"), "%d", &concurrency)
	if concurrency < 1 || concurrency > 8 {
		concurrency = 1
	}
	poll := 3 * time.Second
	if n, e := time.ParseDuration(os.Getenv("WORKER_POLL_SECONDS") + "s"); e == nil && n > 0 {
		poll = n
	}
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				b, status, e := w.request(ctx, "POST", "/internal/jobs/claim", "", map[string]string{"worker_id": w.ID})
				if e == nil && status == 204 {
					b, status, e = w.request(ctx, "POST", "/internal/intakes/claim", "", map[string]string{"worker_id": w.ID})
				}
				if e != nil {
					slog.Warn("claim failed", "error", e)
				} else if status != 204 {
					var claim Claim
					if json.Unmarshal(b, &claim) == nil {
						w.process(ctx, claim, timeout)
					}
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(poll):
				}
			}
		}()
	}
	slog.Info("worker started", "worker_id", w.ID, "concurrency", concurrency)
	wg.Wait()
}
func (w Worker) process(ctx context.Context, c Claim, timeout time.Duration) {
	prefix := "/internal/jobs/"
	if c.Resource == "intakes" {
		prefix = "/internal/intakes/"
	}
	jobctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	done := make(chan struct{})
	go func() {
		tick := time.NewTicker(time.Duration(max(c.LeaseSeconds/3, 1)) * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-jobctx.Done():
				return
			case <-tick.C:
				if _, _, e := w.request(jobctx, "POST", prefix+c.JobID+"/heartbeat", c.Lease, map[string]string{}); e != nil {
					cancel()
					return
				}
			}
		}
	}()
	defer close(done)
	in := extraction.Input{MetricKeys: c.MetricKeys, Operator: c.Operator, JobID: c.JobID, Channel: c.Channel, Scope: c.Scope, Scenario: c.Scenario, Attempt: c.Attempt, Attachments: c.Attachments}
	var err error
	for i := range in.Attachments {
		b, _, e := w.request(jobctx, "GET", prefix+c.JobID+"/attachments/"+in.Attachments[i].ID, c.Lease, nil)
		if e != nil {
			err = e
			break
		}
		in.Attachments[i].Data = b
	}
	var extractor extraction.Extractor
	switch c.Mode {
	case "mock":
		extractor = extraction.MockExtractor{}
	case "openclaw":
		extractor = extraction.OpenClawExtractor{BaseURL: os.Getenv("OPENCLAW_BASE_URL"), Token: os.Getenv("OPENCLAW_GATEWAY_TOKEN"), Agent: os.Getenv("OPENCLAW_AGENT_ID"), Timeout: timeout}
	default:
		extractor = extraction.ManualExtractor{}
	}
	var out extraction.Output
	if err == nil {
		out, err = extractor.Extract(jobctx, in)
	}
	if err == nil {
		_, _, err = w.request(jobctx, "POST", prefix+c.JobID+"/result", c.Lease, out)
	}
	if err != nil {
		message := err.Error()
		w.request(ctx, "POST", prefix+c.JobID+"/failure", c.Lease, map[string]any{"message": message, "transient": strings.HasPrefix(message, "transient:")})
		slog.Warn("job failed", "job_id", c.JobID, "error_class", strings.Split(message, ":")[0])
		return
	}
	slog.Info("job completed", "job_id", c.JobID)
}
