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

-- name: GetDailyUptimeForMonitor :many
SELECT
  date_trunc('day', checked_at)::date AS day,
  count(*)                                    AS total,
  count(*) FILTER (WHERE status = 'up')       AS up_count,
  count(*) FILTER (WHERE status = 'down')     AS down_count,
  count(*) FILTER (WHERE status = 'degraded') AS degraded_count
FROM check_results
WHERE monitor_id = $1
  AND checked_at >= now() - ($2 || ' days')::interval
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_windows mw
    WHERE mw.service_id = (SELECT service_id FROM monitors WHERE id = $1)
      AND mw.archived_at IS NULL
      AND check_results.checked_at BETWEEN mw.starts_at AND mw.ends_at
  )
GROUP BY day
ORDER BY day ASC;

-- name: ListResponseTimes :many
SELECT checked_at, response_time_ms, status FROM check_results
WHERE monitor_id = $1
  AND checked_at >= now() - ($2 || ' hours')::interval
  AND response_time_ms IS NOT NULL
ORDER BY checked_at ASC
LIMIT 500;