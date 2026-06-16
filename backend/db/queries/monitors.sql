-- name: GetMonitorByID :one
SELECT * FROM monitors
WHERE id = $1 AND archived_at IS NULL;

-- name: ListMonitorsByService :many
SELECT * FROM monitors
WHERE service_id = $1 AND archived_at IS NULL
ORDER BY created_at DESC;

-- name: ListDueMonitors :many
SELECT * FROM monitors
WHERE enabled = true
  AND archived_at IS NULL
  AND next_check_at <= now()
ORDER BY next_check_at ASC
LIMIT 100;

-- name: CreateMonitor :one
INSERT INTO monitors (
  service_id, name, type, url, host, port,
  interval_seconds, timeout_ms, retry_count,
  degraded_threshold_ms, http_method, http_expected_status,
  ssl_expiry_threshold_days, keyword_match, keyword_should_exist,
  dns_record_type, dns_expected_value
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
  $13, $14, $15, $16, $17
)
RETURNING *;

-- name: IncrementConsecutiveFailures :one
UPDATE monitors
SET consecutive_failures = consecutive_failures + 1, updated_at = now()
WHERE id = $1
RETURNING consecutive_failures;

-- name: ResetConsecutiveFailures :exec
UPDATE monitors
SET consecutive_failures = 0, updated_at = now()
WHERE id = $1;

-- name: SetNextCheckAt :exec
UPDATE monitors
SET next_check_at = now() + (interval_seconds || ' seconds')::interval
WHERE id = $1;

-- name: UpdateMonitor :one
UPDATE monitors
SET name = $2, url = $3, host = $4, port = $5,
    interval_seconds = $6, timeout_ms = $7, retry_count = $8,
    degraded_threshold_ms = $9, http_method = $10, http_expected_status = $11,
    enabled = $12,
    ssl_expiry_threshold_days = $13, keyword_match = $14, keyword_should_exist = $15,
    dns_record_type = $16, dns_expected_value = $17,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ArchiveMonitor :exec
UPDATE monitors SET archived_at = now(), updated_at = now() WHERE id = $1;