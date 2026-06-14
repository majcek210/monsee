CREATE TABLE incidents (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id   UUID NOT NULL REFERENCES services(id),
  monitor_id   UUID REFERENCES monitors(id),
  title        TEXT NOT NULL,
  severity     TEXT NOT NULL DEFAULT 'high',
  status       TEXT NOT NULL DEFAULT 'open',
  resolved_at  TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON incidents(service_id);
CREATE INDEX ON incidents(monitor_id);
CREATE INDEX ON incidents(status);
