-- name: GetServiceByID :one
SELECT * FROM services
WHERE id = $1 AND archived_at IS NULL;

-- name: ListServices :many
SELECT * FROM services
WHERE archived_at IS NULL
ORDER BY created_at DESC;

-- name: CreateService :one
INSERT INTO services (
    name, description, public_visible, show_uptime,
    dedicated_page_enabled, uptime_range_days, slug, custom_domain, status_override
)
VALUES (
    $1,
    $2,
    COALESCE(sqlc.narg(public_visible), true),
    COALESCE(sqlc.narg(show_uptime), true),
    COALESCE(sqlc.narg(dedicated_page_enabled), false),
    COALESCE(sqlc.narg(uptime_range_days), 90),
    sqlc.narg(slug),
    sqlc.narg(custom_domain),
    sqlc.narg(status_override)
)
RETURNING *;

-- name: UpdateService :one
-- Partial update: any narg left NULL keeps the existing column value.
-- slug/custom_domain/status_override use CASE to allow clearing to NULL.
UPDATE services
SET name                  = COALESCE(sqlc.narg(name), name),
    description           = COALESCE(sqlc.narg(description), description),
    status                = COALESCE(sqlc.narg(status), status),
    public_visible        = COALESCE(sqlc.narg(public_visible), public_visible),
    show_uptime           = COALESCE(sqlc.narg(show_uptime), show_uptime),
    dedicated_page_enabled = COALESCE(sqlc.narg(dedicated_page_enabled), dedicated_page_enabled),
    uptime_range_days     = COALESCE(sqlc.narg(uptime_range_days), uptime_range_days),
    slug                  = CASE WHEN sqlc.narg(slug)::text = '' THEN NULL ELSE COALESCE(sqlc.narg(slug), slug) END,
    custom_domain         = CASE WHEN sqlc.narg(custom_domain)::text = '' THEN NULL ELSE COALESCE(sqlc.narg(custom_domain), custom_domain) END,
    status_override       = CASE WHEN sqlc.narg(status_override)::text = '' THEN NULL ELSE COALESCE(sqlc.narg(status_override), status_override) END
WHERE id = $1
RETURNING *;

-- name: GetServiceBySlug :one
SELECT * FROM services WHERE slug = $1 AND archived_at IS NULL AND dedicated_page_enabled = true;

-- name: GetServiceByCustomDomain :one
SELECT * FROM services WHERE custom_domain = $1 AND archived_at IS NULL AND dedicated_page_enabled = true;

-- name: ListPublic :many
SELECT * FROM services
WHERE archived_at IS NULL AND public_visible = true
ORDER BY created_at DESC;

-- name: ArchiveService :exec
UPDATE services
SET archived_at = now(), slug = NULL, custom_domain = NULL
WHERE id = $1;