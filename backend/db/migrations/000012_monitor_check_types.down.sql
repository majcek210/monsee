ALTER TABLE monitors
  DROP COLUMN IF EXISTS ssl_expiry_threshold_days,
  DROP COLUMN IF EXISTS keyword_match,
  DROP COLUMN IF EXISTS keyword_should_exist,
  DROP COLUMN IF EXISTS dns_record_type,
  DROP COLUMN IF EXISTS dns_expected_value;
