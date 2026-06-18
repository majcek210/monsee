package checks

import (
	"context"
	"fmt"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/netguard"
)

// Run dispatches to the appropriate check implementation based on monitor type.
func Run(ctx context.Context, m *domain.Monitor) Result {
	// SSRF guard: validate the destination is a public address before dispatching.
	switch m.Type {
	case "http", "keyword":
		if m.URL != nil && *m.URL != "" {
			if err := netguard.CheckPublicURL(*m.URL); err != nil {
				return Result{Status: "down", Error: err.Error()}
			}
		}
	case "tcp", "ssl", "dns":
		if m.Host != nil && *m.Host != "" {
			if err := netguard.CheckPublicURL("https://" + *m.Host); err != nil {
				return Result{Status: "down", Error: err.Error()}
			}
		}
	}

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

	case "ssl":
		host := ptrStr(m.Host)
		if host == "" {
			if m.URL != nil {
				host = *m.URL
			}
		}
		var port int32 = 443
		if m.Port != nil {
			port = *m.Port
		}
		return CheckSSL(ctx, SSLCheckParams{
			Host:                host,
			Port:                port,
			TimeoutMs:           int(m.TimeoutMs),
			ExpiryThresholdDays: m.SSLExpiryThresholdDays,
		})

	case "keyword":
		p := KeywordCheckParams{
			URL:       ptrStr(m.URL),
			Method:    ptrStr(m.HTTPMethod),
			TimeoutMs: int(m.TimeoutMs),
			Keyword:   ptrStr(m.KeywordMatch),
			ShouldExist: m.KeywordShouldExist,
		}
		if m.HTTPExpectedStatus != nil {
			p.ExpectedStatus = int(*m.HTTPExpectedStatus)
		}
		if m.DegradedThresholdMs != nil {
			p.DegradedThresholdMs = int(*m.DegradedThresholdMs)
		}
		return CheckKeyword(ctx, p)

	case "dns":
		if m.Host == nil {
			return Result{Status: "down", Error: "dns monitor missing host"}
		}
		return CheckDNS(ctx, DNSCheckParams{
			Host:          *m.Host,
			RecordType:    ptrStr(m.DNSRecordType),
			ExpectedValue: m.DNSExpectedValue,
			TimeoutMs:     int(m.TimeoutMs),
		})

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
