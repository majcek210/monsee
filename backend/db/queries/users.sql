-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND archived_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND archived_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users WHERE archived_at IS NULL ORDER BY created_at;

-- name: UpdateUserRole :one
UPDATE users SET role = $2 WHERE id = $1 AND archived_at IS NULL
RETURNING *;

-- name: CountActiveAdmins :one
SELECT count(*) FROM users WHERE role = 'admin' AND archived_at IS NULL;

-- name: ArchiveUser :exec
UPDATE users SET archived_at = now() WHERE id = $1;
