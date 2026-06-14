// Package netguard helps prevent SSRF by validating that admin-supplied
// destination URLs (webhooks, Discord notification channels, ...) point at
// public hosts rather than internal infrastructure.
package netguard

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

// ErrBlockedTarget is returned when a URL points at a non-public address.
var ErrBlockedTarget = errors.New("target address is not allowed")

// CheckPublicURL returns an error if rawURL is not an http(s) URL, or if its
// host resolves to a loopback, private, link-local, unspecified, or multicast
// address (e.g. 127.0.0.1, 10.0.0.0/8, 169.254.169.254 cloud metadata, ::1).
//
// This performs a DNS lookup for hostnames, so it reflects where the request
// would actually go right now. Use it at the delivery boundary (immediately
// before contacting the destination) so DNS-rebinding after configuration
// can't reach internal hosts.
func CheckPublicURL(rawURL string) error {
	host, err := checkScheme(rawURL)
	if err != nil {
		return err
	}

	ips, err := resolveIPs(host)
	if err != nil {
		return fmt.Errorf("resolve host %q: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("resolve host %q: no addresses found", host)
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: %s resolves to %s", ErrBlockedTarget, host, ip)
		}
	}
	return nil
}

// CheckPublicURLSyntax validates scheme and host like CheckPublicURL and
// blocks IP-literal hosts (e.g. http://127.0.0.1/, http://169.254.169.254/)
// but performs no DNS lookup for hostnames.
//
// Use this for config-time validation (Create/Update) to give immediate
// feedback on obviously-bad URLs without making admin writes depend on DNS.
// CheckPublicURL is still enforced at delivery time as the real security
// boundary — a hostname that looks fine here can still be blocked later if
// it resolves to a private address.
func CheckPublicURLSyntax(rawURL string) error {
	host, err := checkScheme(rawURL)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil && isBlockedIP(ip) {
		return fmt.Errorf("%w: %s", ErrBlockedTarget, host)
	}
	return nil
}

// checkScheme validates that rawURL is an http(s) URL with a non-empty host
// and returns that host.
func checkScheme(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("url scheme must be http or https")
	}

	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("url has no host")
	}
	return host, nil
}

func resolveIPs(host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	return net.LookupIP(host)
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}
