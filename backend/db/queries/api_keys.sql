-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, name, key_hash, prefix)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND archived_at IS NULL;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys WHERE id = $1 AND archived_at IS NULL;

-- name: ListAPIKeysByUser :many
SELECT * FROM api_keys WHERE user_id = $1 AND archived_at IS NULL ORDER BY created_at DESC;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used = now() WHERE id = $1;

-- name: ArchiveAPIKey :exec
UPDATE api_keys SET archived_at = now() WHERE id = $1;
