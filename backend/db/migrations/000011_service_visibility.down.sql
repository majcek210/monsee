DROP INDEX IF EXISTS services_slug_unique;
DROP INDEX IF EXISTS services_custom_domain_unique;

ALTER TABLE services
  DROP COLUMN IF EXISTS public_visible,
  DROP COLUMN IF EXISTS show_uptime,
  DROP COLUMN IF EXISTS dedicated_page_enabled,
  DROP COLUMN IF EXISTS slug,
  DROP COLUMN IF EXISTS custom_domain,
  DROP COLUMN IF EXISTS uptime_range_days;
