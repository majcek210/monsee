package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/majcek210/monsee/pkg/netguard"
)

type DeliveryResult struct {
	StatusCode int
	DurationMs int32
	Err        string
}

type EventPayload struct {
	Event     string         `json:"event"`
	Timestamp string         `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// Deliver POSTs the event payload to url with optional HMAC-SHA256 signature.
// Retries up to maxRetries on failure.
func Deliver(ctx context.Context, url, secret string, payload EventPayload, maxRetries int) DeliveryResult {
	body, _ := json.Marshal(payload)

	var result DeliveryResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		result = doDeliver(ctx, url, secret, body)
		if result.Err == "" {
			return result
		}
	}
	return result
}

func doDeliver(ctx context.Context, url, secret string, body []byte) DeliveryResult {
	// SSRF guard: re-checked at delivery time (not just at create/update) so
	// DNS rebinding after the webhook was configured can't reach internal hosts.
	if err := netguard.CheckPublicURL(url); err != nil {
		return DeliveryResult{Err: "blocked target: " + err.Error()}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return DeliveryResult{Err: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "status-monitor/1.0")

	if secret != "" {
		sig := signHMAC(secret, body)
		req.Header.Set("X-Webhook-Signature", "sha256="+sig)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := int32(time.Since(start).Milliseconds())

	if err != nil {
		return DeliveryResult{DurationMs: elapsed, Err: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DeliveryResult{
			StatusCode: resp.StatusCode,
			DurationMs: elapsed,
			Err:        fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}
	return DeliveryResult{StatusCode: resp.StatusCode, DurationMs: elapsed}
}

func signHMAC(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
