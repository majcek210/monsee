-- name: InsertCheckResult :one
INSERT INTO check_results (monitor_id, status, response_time_ms, error)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestCheckForMonitor :one
SELECT * FROM check_results
WHERE monitor_id = $1
ORDER BY checked_at DESC
LIMIT 1;

-- name: ListCheckResultsByMonitor :many
SELECT * FROM check_results
WHERE monitor_id = $1
ORDER BY checked_at DESC
LIMIT $2;