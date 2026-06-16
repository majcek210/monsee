ALTER TABLE users
  ADD COLUMN totp_secret       TEXT,
  ADD COLUMN totp_enabled      BOOL NOT NULL DEFAULT false,
  ADD COLUMN totp_backup_codes TEXT[];
