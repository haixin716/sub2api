-- Keep f_dev_3.0 usage-log writes compatible after migration 046 introduced
-- client_request_id and migration 182 restored request_id deduplication.
--
-- The f_dev_3.0 repository does not write client_request_id to usage_logs and
-- deliberately stores an empty RequestID as SQL NULL for legacy/test callers.
-- PostgreSQL unique indexes still deduplicate non-NULL request IDs while
-- allowing multiple NULL values, so both columns must remain nullable.

ALTER TABLE usage_logs
    ALTER COLUMN client_request_id DROP NOT NULL;

ALTER TABLE usage_logs
    ALTER COLUMN request_id DROP NOT NULL;
