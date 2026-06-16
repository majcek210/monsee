ALTER TABLE monitors
  ADD COLUMN ssl_expiry_threshold_days INT NOT NULL DEFAULT 14,
  ADD COLUMN keyword_match             TEXT,
  ADD COLUMN keyword_should_exist      BOOL NOT NULL DEFAULT true,
  ADD COLUMN dns_record_type           TEXT,
  ADD COLUMN dns_expected_value        TEXT;
