-- 修复唯一约束问题
-- 为 usage_logs 和 request_logs 添加 (client_request_id, api_key_id) 唯一约束

BEGIN;

-- 1. 为 usage_logs 添加唯一约束
-- 注意：如果表中已有重复数据，此操作会失败
-- 可以先查询检查：SELECT client_request_id, api_key_id, COUNT(*) FROM usage_logs GROUP BY client_request_id, api_key_id HAVING COUNT(*) > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_logs_client_request_id_api_key_unique
    ON usage_logs (client_request_id, api_key_id);

-- 2. 为 request_logs 添加唯一约束
-- 注意：request_logs 可能不需要这个约束，取决于业务逻辑
-- 如果一个 client_request_id 可能对应多条 request_log（比如重试场景），则不应该添加
CREATE UNIQUE INDEX IF NOT EXISTS idx_request_logs_client_request_id_api_key_unique
    ON request_logs (client_request_id, api_key_id);

-- 3. (可选) 如果不再需要旧的 request_id 唯一约束，可以删除
-- DROP INDEX IF EXISTS idx_usage_logs_request_id_api_key_unique;

COMMIT;

-- 验证
-- \d usage_logs
-- \d request_logs
