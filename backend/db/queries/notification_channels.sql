-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (name, type, config)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetNotificationChannelByID :one
SELECT * FROM notification_channels WHERE id = $1 AND archived_at IS NULL;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels WHERE archived_at IS NULL ORDER BY created_at;

-- name: UpdateNotificationChannel :one
-- Partial update: any narg left NULL keeps the existing column value, so the
-- handler can omit name/config/enabled independently (e.g. the "toggle
-- enabled" switch sends only enabled).
UPDATE notification_channels
SET name    = COALESCE(sqlc.narg(name), name),
    config  = COALESCE(sqlc.narg(config), config),
    enabled = COALESCE(sqlc.narg(enabled), enabled)
WHERE id = $1
RETURNING *;

-- name: ArchiveNotificationChannel :exec
UPDATE notification_channels SET archived_at = now() WHERE id = $1;

-- name: ListEnabledNotificationChannels :many
SELECT * FROM notification_channels WHERE enabled = true AND archived_at IS NULL;
