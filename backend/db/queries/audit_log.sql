-- name: InsertAuditLog :one
INSERT INTO audit_log (user_id, action, resource, resource_id, ip, user_agent, diff)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLog :many
SELECT * FROM audit_log
WHERE (sqlc.narg(resource)::text IS NULL OR resource = sqlc.narg(resource))
  AND (sqlc.narg(filter_user_id)::uuid IS NULL OR user_id = sqlc.narg(filter_user_id))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLog :one
SELECT count(*) FROM audit_log
WHERE (sqlc.narg(resource)::text IS NULL OR resource = sqlc.narg(resource))
  AND (sqlc.narg(filter_user_id)::uuid IS NULL OR user_id = sqlc.narg(filter_user_id));

-- name: ListAuditLogByResource :many
SELECT * FROM audit_log
WHERE resource = $1 AND resource_id = $2
ORDER BY created_at DESC;
