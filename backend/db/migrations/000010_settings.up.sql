CREATE TABLE settings (
  id                    INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  site_title            TEXT NOT NULL DEFAULT 'monsee',
  logo_url              TEXT NOT NULL DEFAULT '',
  public_status_enabled BOOL NOT NULL DEFAULT true,
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO settings (id) VALUES (1);
