CREATE TABLE webhooks (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  url         TEXT NOT NULL,
  secret      TEXT,
  events      TEXT[] NOT NULL DEFAULT '{}',
  enabled     BOOL NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ
);

CREATE TABLE webhook_logs (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  webhook_id   UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
  event        TEXT NOT NULL,
  status_code  INT,
  error        TEXT,
  duration_ms  INT,
  delivered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON webhook_logs(webhook_id, delivered_at DESC);
