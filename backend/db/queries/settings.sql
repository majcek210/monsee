-- name: GetSettings :one
SELECT * FROM settings WHERE id = 1;

-- name: UpdateSettings :one
UPDATE settings
SET site_title             = COALESCE(sqlc.narg(site_title), site_title),
    logo_url               = COALESCE(sqlc.narg(logo_url), logo_url),
    public_status_enabled  = COALESCE(sqlc.narg(public_status_enabled), public_status_enabled),
    custom_domains_enabled = COALESCE(sqlc.narg(custom_domains_enabled), custom_domains_enabled),
    updated_at             = now()
WHERE id = 1
RETURNING *;
