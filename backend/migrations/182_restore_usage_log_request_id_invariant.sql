-- f_dev_1.0 migration 046 temporarily changed usage_logs to use a separate
-- client_request_id key. f_dev_3.0 already treats request_id as its stable,
-- idempotent request key, so restore that invariant after all legacy migrations
-- have run while retaining the extra column for backwards compatibility.

UPDATE usage_logs
SET request_id = COALESCE(NULLIF(request_id, ''), NULLIF(client_request_id, ''), 'legacy-' || id::text)
WHERE request_id IS NULL OR request_id = '';

ALTER TABLE usage_logs
    ALTER COLUMN request_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_logs_request_id_api_key_unique
    ON usage_logs(request_id, api_key_id);
