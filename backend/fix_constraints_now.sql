-- 立即修复约束问题的脚本
-- 这个脚本可以直接在当前数据库上执行，无需回滚

BEGIN;

-- ============================================================================
-- Step 1: 修复 usage_logs 表
-- ============================================================================

-- 1.1 删除旧的 (request_id, api_key_id) 唯一约束
-- 这个约束会导致插入失败，因为 request_id 可能重复
DROP INDEX IF EXISTS idx_usage_logs_request_id_api_key_unique;

-- 1.2 添加新的 (client_request_id, api_key_id) 唯一约束
-- 这是 ON CONFLICT 子句所需要的
CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_logs_client_request_id_api_key_unique
    ON usage_logs (client_request_id, api_key_id);

-- ============================================================================
-- Step 2: 修复 request_logs 表
-- ============================================================================

-- 2.1 添加 (client_request_id, api_key_id) 唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_request_logs_client_request_id_api_key_unique
    ON request_logs (client_request_id, api_key_id);

COMMIT;

-- ============================================================================
-- 验证
-- ============================================================================

-- 查看 usage_logs 的唯一约束
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'usage_logs'
  AND (indexname LIKE '%unique%' OR indexname LIKE '%client_request%')
ORDER BY indexname;

-- 查看 request_logs 的唯一约束
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'request_logs'
  AND (indexname LIKE '%unique%' OR indexname LIKE '%client_request%')
ORDER BY indexname;

-- 测试 ON CONFLICT 是否工作
-- 注意：需要替换 user_id, api_key_id, account_id 为实际存在的值
-- INSERT INTO usage_logs (
--     user_id, api_key_id, account_id,
--     client_request_id, model, created_at
-- ) VALUES (
--     1, 1, 1,
--     'test-conflict-' || gen_random_uuid()::text,
--     'claude-3-5-sonnet-20241022',
--     NOW()
-- ) ON CONFLICT (client_request_id, api_key_id) DO NOTHING;
