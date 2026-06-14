package checks

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Result struct {
	Status         string // "up", "down", "degraded"
	ResponseTimeMs int
	Error          string
}

type HTTPCheckParams struct {
	URL                 string
	Method              string
	TimeoutMs           int
	ExpectedStatus      int
	DegradedThresholdMs int
}

func CheckHTTP(ctx context.Context, p HTTPCheckParams) Result {
	method := p.Method
	if method == "" {
		method = "GET"
	}

	client := &http.Client{
		Timeout: time.Duration(p.TimeoutMs) * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, method, p.URL, nil)
	if err != nil {
		return Result{Status: "down", Error: err.Error()}
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{Status: "down", Error: err.Error(), ResponseTimeMs: elapsed}
	}
	defer resp.Body.Close()

	expected := p.ExpectedStatus
	if expected == 0 {
		expected = 200
	}
	if resp.StatusCode != expected {
		return Result{
			Status:         "down",
			Error:          fmt.Sprintf("got %d, expected %d", resp.StatusCode, expected),
			ResponseTimeMs: elapsed,
		}
	}

	if p.DegradedThresholdMs > 0 && elapsed > p.DegradedThresholdMs {
		return Result{Status: "degraded", ResponseTimeMs: elapsed}
	}

	return Result{Status: "up", ResponseTimeMs: elapsed}
}
