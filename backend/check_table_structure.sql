-- 检查 usage_logs 表结构
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'usage_logs'
AND column_name IN ('client_request_id', 'request_id')
ORDER BY ordinal_position;

-- 检查 request_logs 表结构
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'request_logs'
AND column_name IN ('client_request_id', 'request_id')
ORDER BY ordinal_position;

-- 检查约束
SELECT conname, contype, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'usage_logs'::regclass
AND conname LIKE '%client_request%';

SELECT conname, contype, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'request_logs'::regclass
AND conname LIKE '%client_request%';

-- 检查索引
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'usage_logs'
AND indexname LIKE '%client_request%';

SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'request_logs'
AND indexname LIKE '%client_request%';
