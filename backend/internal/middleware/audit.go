package middleware

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// AuditLogger is the interface the audit middleware uses to write entries.
type AuditLogger interface {
	Log(ctx context.Context, entry AuditEntry) error
}

type AuditEntry struct {
	UserID     string
	Action     string // "create" | "update" | "archive" | "delete"
	Resource   string // "monitor" | "service" | "incident" ...
	ResourceID string
	IP         string
	UserAgent  string
	// Diff contains field names only — NEVER decrypted values
	Diff map[string]any
}

// Audit logs every successful admin write (POST/PATCH/DELETE) to the audit log.
// It infers resource and action from the request path and method.
func Audit(logger AuditLogger) fiber.Handler {
	return func(c fiber.Ctx) error {
		method := c.Method()

		// Capture field names from the request body before the handler runs.
		// Diff records WHICH fields were submitted, never their values, so
		// encrypted/secret fields (webhook URL, notification config, ...)
		// never leak into the audit trail.
		var diff map[string]any
		if method == "POST" || method == "PATCH" {
			if fields := auditFieldNames(c.Body()); len(fields) > 0 {
				diff = map[string]any{"fields": fields}
			}
		}

		if err := c.Next(); err != nil {
			return err
		}

		if method != "POST" && method != "PATCH" && method != "DELETE" {
			return nil
		}

		// Only log 2xx responses
		status := c.Response().StatusCode()
		if status < 200 || status >= 300 {
			return nil
		}

		action := methodToAction(method)
		resource, resourceID := parsePath(c.Path())

		entry := AuditEntry{
			UserID:     UserIDFromCtx(c),
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			IP:         c.IP(),
			UserAgent:  string(c.Request().Header.UserAgent()),
			Diff:       diff,
		}

		// Log async so we never block the response
		go func() {
			_ = logger.Log(context.Background(), entry)
		}()

		return nil
	}
}

// auditFieldNames returns the sorted top-level JSON field names present in
// body, or nil if body is empty/not a JSON object. Used to record WHAT was
// changed in a request without ever recording field values.
func auditFieldNames(body []byte) []string {
	if len(body) == 0 {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	fields := make([]string, 0, len(raw))
	for k := range raw {
		fields = append(fields, k)
	}
	sort.Strings(fields)
	return fields
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PATCH":
		return "update"
	case "DELETE":
		return "archive"
	}
	return "unknown"
}

// parsePath extracts resource and optional resource ID from /admin/<resource>[/<id>]
func parsePath(path string) (resource, resourceID string) {
	parts := strings.Split(strings.TrimPrefix(path, "/admin/"), "/")
	if len(parts) > 0 {
		resource = parts[0]
	}
	if len(parts) > 1 {
		resourceID = parts[1]
	}
	return
}
