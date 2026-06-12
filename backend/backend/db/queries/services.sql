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
UPDATE services
SET name = $2, description = $3, status = $4
WHERE id = $1
RETURNING *;

-- name: ArchiveService :exec
UPDATE services
SET archived_at = now()
WHERE id = $1;