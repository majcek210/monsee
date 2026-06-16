package checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type KeywordCheckParams struct {
	URL                 string
	Method              string
	TimeoutMs           int
	ExpectedStatus      int
	Keyword             string
	ShouldExist         bool
	DegradedThresholdMs int
}

func CheckKeyword(ctx context.Context, p KeywordCheckParams) Result {
	if p.Method == "" {
		p.Method = "GET"
	}
	if p.ExpectedStatus == 0 {
		p.ExpectedStatus = 200
	}
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, nil)
	if err != nil {
		return Result{Status: "down", Error: err.Error()}
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: "failed to read body: " + err.Error()}
	}

	found := strings.Contains(string(body), p.Keyword)
	if found != p.ShouldExist {
		msg := fmt.Sprintf("keyword %q not found", p.Keyword)
		if !p.ShouldExist {
			msg = fmt.Sprintf("keyword %q found but should not exist", p.Keyword)
		}
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: msg}
	}

	if resp.StatusCode != p.ExpectedStatus {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: fmt.Sprintf("expected status %d, got %d", p.ExpectedStatus, resp.StatusCode)}
	}

	if p.DegradedThresholdMs > 0 && elapsed >= p.DegradedThresholdMs {
		return Result{Status: "degraded", ResponseTimeMs: elapsed}
	}
	return Result{Status: "up", ResponseTimeMs: elapsed}
}
