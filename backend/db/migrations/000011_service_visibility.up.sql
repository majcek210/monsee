ALTER TABLE services
  ADD COLUMN public_visible         BOOL NOT NULL DEFAULT true,
  ADD COLUMN show_uptime            BOOL NOT NULL DEFAULT true,
  ADD COLUMN dedicated_page_enabled BOOL NOT NULL DEFAULT false,
  ADD COLUMN slug                   TEXT,
  ADD COLUMN custom_domain          TEXT,
  ADD COLUMN uptime_range_days      INT NOT NULL DEFAULT 90;

CREATE UNIQUE INDEX services_slug_unique ON services (slug) WHERE slug IS NOT NULL;
CREATE UNIQUE INDEX services_custom_domain_unique ON services (custom_domain) WHERE custom_domain IS NOT NULL;
