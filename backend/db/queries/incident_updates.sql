-- name: CreateIncidentUpdate :one
INSERT INTO incident_updates (incident_id, status, message)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListIncidentUpdates :many
SELECT * FROM incident_updates
WHERE incident_id = $1
ORDER BY created_at ASC;
