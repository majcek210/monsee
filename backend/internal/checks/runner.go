package checks

import (
	"context"
	"fmt"

	"github.com/majcek210/monsee/internal/domain"
)

// Run dispatches to the appropriate check implementation based on monitor type.
func Run(ctx context.Context, m *domain.Monitor) Result {
	switch m.Type {
	case "http":
		p := HTTPCheckParams{
			URL:       ptrStr(m.URL),
			Method:    ptrStr(m.HTTPMethod),
			TimeoutMs: int(m.TimeoutMs),
		}
		if m.DegradedThresholdMs != nil {
			p.DegradedThresholdMs = int(*m.DegradedThresholdMs)
		}
		if m.HTTPExpectedStatus != nil {
			p.ExpectedStatus = int(*m.HTTPExpectedStatus)
		}
		return CheckHTTP(ctx, p)

	case "tcp":
		if m.Host == nil || m.Port == nil {
			return Result{Status: "down", Error: "tcp monitor missing host or port"}
		}
		p := TCPCheckParams{
			Host:      *m.Host,
			Port:      *m.Port,
			TimeoutMs: int(m.TimeoutMs),
		}
		if m.DegradedThresholdMs != nil {
			p.DegradedThresholdMs = int(*m.DegradedThresholdMs)
		}
		return CheckTCP(ctx, p)

	default:
		return Result{Status: "down", Error: fmt.Sprintf("unknown monitor type: %s", m.Type)}
	}
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
