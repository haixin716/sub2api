-- 047_add_request_log_cleanup_tasks.sql
-- 请求记录清理任务表

CREATE TABLE IF NOT EXISTS request_log_cleanup_tasks (
    id BIGSERIAL PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    filters JSONB NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    deleted_rows BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    canceled_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    canceled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_log_cleanup_tasks_status_created_at
    ON request_log_cleanup_tasks(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_log_cleanup_tasks_created_at
    ON request_log_cleanup_tasks(created_at DESC);
