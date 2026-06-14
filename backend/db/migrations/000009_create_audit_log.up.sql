CREATE TABLE audit_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID REFERENCES users(id),
  action      TEXT NOT NULL,
  resource    TEXT NOT NULL,
  resource_id TEXT,
  ip          TEXT,
  user_agent  TEXT,
  diff        JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON audit_log(user_id);
CREATE INDEX ON audit_log(resource, resource_id);
CREATE INDEX ON audit_log(created_at DESC);
