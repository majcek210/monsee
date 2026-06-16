package checks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type SSLCheckParams struct {
	Host                string
	Port                int32
	TimeoutMs           int
	ExpiryThresholdDays int32
}

func CheckSSL(ctx context.Context, p SSLCheckParams) Result {
	if p.Port == 0 {
		p.Port = 443
	}
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	start := time.Now()
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: p.Host,
	})
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: err.Error()}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: "no certificates found"}
	}

	expiry := certs[0].NotAfter
	daysUntil := int32(time.Until(expiry).Hours() / 24)

	if daysUntil < 0 {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: "certificate has expired"}
	}
	if daysUntil <= p.ExpiryThresholdDays {
		return Result{
			Status:         "degraded",
			ResponseTimeMs: elapsed,
			Error:          fmt.Sprintf("certificate expires in %d days", daysUntil),
		}
	}
	return Result{Status: "up", ResponseTimeMs: elapsed}
}
