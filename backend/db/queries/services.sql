-- name: GetServiceByID :one
SELECT * FROM services
WHERE id = $1 AND archived_at IS NULL;

-- name: ListServices :many
SELECT * FROM services
WHERE archived_at IS NULL
ORDER BY created_at DESC;

-- name: CreateService :one
INSERT INTO services (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateService :one
-- Partial update: any narg left NULL keeps the existing column value, so the
-- handler can omit name/description/status independently.
UPDATE services
SET name        = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    status      = COALESCE(sqlc.narg(status), status)
WHERE id = $1
RETURNING *;

-- name: ArchiveService :exec
UPDATE services
SET archived_at = now()
WHERE id = $1;