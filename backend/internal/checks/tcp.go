package checks

import (
	"context"
	"fmt"
	"net"
	"time"
)

type TCPCheckParams struct {
	Host                string
	Port                int32
	TimeoutMs           int
	DegradedThresholdMs int
}

func CheckTCP(ctx context.Context, p TCPCheckParams) Result {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond

	d := net.Dialer{Timeout: timeout}

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", addr)
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{Status: "down", Error: err.Error(), ResponseTimeMs: elapsed}
	}
	conn.Close()

	if p.DegradedThresholdMs > 0 && elapsed > p.DegradedThresholdMs {
		return Result{Status: "degraded", ResponseTimeMs: elapsed}
	}

	return Result{Status: "up", ResponseTimeMs: elapsed}
}
