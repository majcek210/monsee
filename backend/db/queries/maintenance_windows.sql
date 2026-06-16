-- name: CreateMaintenanceWindow :one
INSERT INTO maintenance_windows (service_id, title, description, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMaintenanceWindowByID :one
SELECT * FROM maintenance_windows WHERE id = $1 AND archived_at IS NULL;

-- name: ListMaintenanceWindowsByService :many
SELECT * FROM maintenance_windows
WHERE service_id = $1 AND archived_at IS NULL
ORDER BY starts_at DESC;

-- name: ListActiveMaintenanceWindows :many
SELECT * FROM maintenance_windows
WHERE archived_at IS NULL
  AND starts_at <= now()
  AND ends_at >= now()
ORDER BY starts_at ASC;

-- name: ListActiveForService :many
SELECT * FROM maintenance_windows
WHERE service_id = $1
  AND archived_at IS NULL
  AND starts_at <= now()
  AND ends_at >= now()
ORDER BY starts_at ASC;

-- name: IsMaintenanceActiveForService :one
SELECT EXISTS(
  SELECT 1 FROM maintenance_windows
  WHERE service_id = $1
    AND archived_at IS NULL
    AND starts_at <= now()
    AND ends_at >= now()
) AS active;

-- name: UpdateMaintenanceWindow :one
UPDATE maintenance_windows
SET title       = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    starts_at   = COALESCE(sqlc.narg(starts_at), starts_at),
    ends_at     = COALESCE(sqlc.narg(ends_at), ends_at)
WHERE id = $1 AND archived_at IS NULL
RETURNING *;

-- name: ArchiveMaintenanceWindow :exec
UPDATE maintenance_windows SET archived_at = now() WHERE id = $1;
