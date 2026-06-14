package netguard

import (
	"errors"
	"testing"
)

func TestCheckPublicURLBlocksPrivateAndLoopback(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1/webhook",
		"http://127.0.0.1:8080/",
		"http://localhost/",
		"http://10.0.0.5/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data/", // cloud metadata
		"http://0.0.0.0/",
		"http://[::1]/",
		"http://[fe80::1]/",
		"http://[fc00::1]/",
	}

	for _, raw := range blocked {
		if err := CheckPublicURL(raw); err == nil {
			t.Errorf("CheckPublicURL(%q) = nil, want error", raw)
		} else if !errors.Is(err, ErrBlockedTarget) {
			t.Errorf("CheckPublicURL(%q) = %v, want ErrBlockedTarget", raw, err)
		}
	}
}

func TestCheckPublicURLAllowsPublicIPLiteral(t *testing.T) {
	if err := CheckPublicURL("https://8.8.8.8/webhook"); err != nil {
		t.Errorf("CheckPublicURL(public IP) = %v, want nil", err)
	}
}

func TestCheckPublicURLRejectsBadScheme(t *testing.T) {
	for _, raw := range []string{
		"ftp://example.com/",
		"file:///etc/passwd",
		"gopher://example.com/",
		"not a url",
	} {
		if err := CheckPublicURL(raw); err == nil {
			t.Errorf("CheckPublicURL(%q) = nil, want error", raw)
		}
	}
}

func TestCheckPublicURLRejectsNoHost(t *testing.T) {
	if err := CheckPublicURL("http:///path"); err == nil {
		t.Error("CheckPublicURL with empty host = nil, want error")
	}
}

func TestCheckPublicURLSyntaxBlocksIPLiterals(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1/webhook",
		"http://127.0.0.1:8080/",
		"http://10.0.0.5/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data/", // cloud metadata
		"http://0.0.0.0/",
		"http://[::1]/",
		"http://[fe80::1]/",
		"http://[fc00::1]/",
	}

	for _, raw := range blocked {
		if err := CheckPublicURLSyntax(raw); err == nil {
			t.Errorf("CheckPublicURLSyntax(%q) = nil, want error", raw)
		} else if !errors.Is(err, ErrBlockedTarget) {
			t.Errorf("CheckPublicURLSyntax(%q) = %v, want ErrBlockedTarget", raw, err)
		}
	}
}

// CheckPublicURLSyntax must not perform DNS lookups: hostnames are allowed
// through regardless of what they resolve to, since the real enforcement
// happens at delivery time via CheckPublicURL.
func TestCheckPublicURLSyntaxAllowsHostnamesWithoutDNS(t *testing.T) {
	allowed := []string{
		"https://example.com/webhook",
		"https://discord.com/api/webhooks/x",
		"http://localhost/",                 // hostname, not an IP literal
		"http://internal.invalid-tld-zzz/",  // would fail DNS, but syntax check doesn't resolve
	}

	for _, raw := range allowed {
		if err := CheckPublicURLSyntax(raw); err != nil {
			t.Errorf("CheckPublicURLSyntax(%q) = %v, want nil", raw, err)
		}
	}
}

func TestCheckPublicURLSyntaxRejectsBadScheme(t *testing.T) {
	for _, raw := range []string{
		"ftp://example.com/",
		"file:///etc/passwd",
		"not a url",
		"http:///path",
	} {
		if err := CheckPublicURLSyntax(raw); err == nil {
			t.Errorf("CheckPublicURLSyntax(%q) = nil, want error", raw)
		}
	}
}
