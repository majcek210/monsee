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

-- name: UpdateUserEmail :one
UPDATE users SET email = $2 WHERE id = $1 AND archived_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1 AND archived_at IS NULL;

-- name: CountActiveAdmins :one
SELECT count(*) FROM users WHERE role = 'admin' AND archived_at IS NULL;

-- name: ArchiveUser :exec
UPDATE users SET archived_at = now() WHERE id = $1;

-- name: GetTOTPByUserID :one
SELECT totp_secret, totp_enabled, totp_backup_codes FROM users WHERE id = $1;

-- name: SetTOTPSecret :exec
UPDATE users SET totp_secret = $2 WHERE id = $1;

-- name: EnableTOTP :exec
UPDATE users SET totp_enabled = true, totp_backup_codes = $2 WHERE id = $1;

-- name: DisableTOTP :exec
UPDATE users SET totp_enabled = false, totp_secret = NULL, totp_backup_codes = NULL WHERE id = $1;

-- name: RemoveBackupCode :exec
UPDATE users SET totp_backup_codes = array_remove(totp_backup_codes, $2) WHERE id = $1;

-- name: ConsumeBackupCode :execrows
-- Atomically removes a backup code only if it is present in the array.
-- Returns 1 if consumed, 0 if the code was not found (wrong or already used).
UPDATE users
SET totp_backup_codes = array_remove(totp_backup_codes, $2)
WHERE id = $1 AND $2 = ANY(totp_backup_codes);
