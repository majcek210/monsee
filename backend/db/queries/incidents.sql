-- name: CreateIncident :one
INSERT INTO incidents (service_id, monitor_id, title, severity)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetIncidentByID :one
SELECT * FROM incidents WHERE id = $1;

-- name: GetOpenIncidentForMonitor :one
SELECT * FROM incidents
WHERE monitor_id = $1 AND status = 'open'
LIMIT 1;

-- name: ListIncidentsByService :many
SELECT * FROM incidents WHERE service_id = $1 ORDER BY created_at DESC;

-- name: ListAllIncidents :many
SELECT * FROM incidents ORDER BY created_at DESC;

-- name: ResolveIncident :one
UPDATE incidents
SET status = 'resolved', resolved_at = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateIncident :one
UPDATE incidents
SET title = $2, severity = $3, status = $4, updated_at = now()
WHERE id = $1
RETURNING *;
