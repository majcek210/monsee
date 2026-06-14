CREATE TABLE check_results (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  monitor_id       UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
  status           TEXT NOT NULL,
  response_time_ms INT,
  error            TEXT,
  checked_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON check_results(monitor_id, checked_at DESC);