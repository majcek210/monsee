CREATE TABLE api_keys (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id),
  name        TEXT NOT NULL,
  key_hash    TEXT NOT NULL UNIQUE,
  prefix      TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used   TIMESTAMPTZ,
  archived_at TIMESTAMPTZ
);
