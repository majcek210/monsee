-- name: InsertAuditLog :one
INSERT INTO audit_log (user_id, action, resource, resource_id, ip, user_agent, diff)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLog :many
SELECT * FROM audit_log ORDER BY created_at DESC LIMIT $1;

-- name: ListAuditLogByResource :many
SELECT * FROM audit_log
WHERE resource = $1 AND resource_id = $2
ORDER BY created_at DESC;
