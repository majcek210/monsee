-- name: CreateWebhook :one
INSERT INTO webhooks (name, url, secret, events)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhookByID :one
SELECT * FROM webhooks WHERE id = $1 AND archived_at IS NULL;

-- name: ListWebhooks :many
SELECT * FROM webhooks WHERE archived_at IS NULL ORDER BY created_at;

-- name: UpdateWebhook :one
-- Partial update: any narg left NULL keeps the existing column value, so the
-- handler can omit name/url/secret/events/enabled independently (e.g. the
-- "toggle enabled" switch sends only enabled).
UPDATE webhooks
SET name    = COALESCE(sqlc.narg(name), name),
    url     = COALESCE(sqlc.narg(url), url),
    secret  = COALESCE(sqlc.narg(secret), secret),
    events  = COALESCE(sqlc.narg(events), events),
    enabled = COALESCE(sqlc.narg(enabled), enabled)
WHERE id = $1
RETURNING *;

-- name: ArchiveWebhook :exec
UPDATE webhooks SET archived_at = now() WHERE id = $1;

-- name: ListWebhooksByEvent :many
SELECT * FROM webhooks
WHERE enabled = true AND archived_at IS NULL AND $1::text = ANY(events);

-- name: InsertWebhookLog :one
INSERT INTO webhook_logs (webhook_id, event, status_code, error, duration_ms)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListWebhookLogs :many
SELECT * FROM webhook_logs WHERE webhook_id = $1 ORDER BY delivered_at DESC LIMIT $2;
