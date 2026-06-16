package checks

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type DNSCheckParams struct {
	Host          string
	RecordType    string
	ExpectedValue *string
	TimeoutMs     int
}

func CheckDNS(ctx context.Context, p DNSCheckParams) Result {
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolver := &net.Resolver{}
	start := time.Now()

	var results []string
	var lookupErr error

	switch strings.ToUpper(p.RecordType) {
	case "A", "AAAA", "":
		addrs, err := resolver.LookupHost(rctx, p.Host)
		results = addrs
		lookupErr = err
	case "CNAME":
		cname, err := resolver.LookupCNAME(rctx, p.Host)
		results = []string{cname}
		lookupErr = err
	case "MX":
		mx, err := resolver.LookupMX(rctx, p.Host)
		for _, r := range mx {
			results = append(results, r.Host)
		}
		lookupErr = err
	case "TXT":
		txt, err := resolver.LookupTXT(rctx, p.Host)
		results = txt
		lookupErr = err
	case "NS":
		ns, err := resolver.LookupNS(rctx, p.Host)
		for _, r := range ns {
			results = append(results, r.Host)
		}
		lookupErr = err
	default:
		return Result{Status: "down", Error: fmt.Sprintf("unsupported record type: %s", p.RecordType)}
	}

	elapsed := int(time.Since(start).Milliseconds())

	if lookupErr != nil {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: lookupErr.Error()}
	}
	if len(results) == 0 {
		return Result{Status: "down", ResponseTimeMs: elapsed, Error: "no records found"}
	}

	if p.ExpectedValue != nil && *p.ExpectedValue != "" {
		found := false
		for _, r := range results {
			if strings.Contains(r, *p.ExpectedValue) {
				found = true
				break
			}
		}
		if !found {
			return Result{Status: "down", ResponseTimeMs: elapsed, Error: fmt.Sprintf("expected value %q not found in results", *p.ExpectedValue)}
		}
	}

	return Result{Status: "up", ResponseTimeMs: elapsed}
}
