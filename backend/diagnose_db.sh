#!/bin/bash

# 诊断数据库问题的脚本

echo "=== 检查 usage_logs 表结构 ==="
psql -U "$PGUSER" -d "$PGDATABASE" -c "\d usage_logs" 2>&1 | grep -E "client_request_id|request_id"

echo ""
echo "=== 检查 request_logs 表结构 ==="
psql -U "$PGUSER" -d "$PGDATABASE" -c "\d request_logs" 2>&1 | grep -E "client_request_id|request_id"

echo ""
echo "=== 检查最近的 usage_logs 记录 ==="
psql -U "$PGUSER" -d "$PGDATABASE" -c "SELECT id, user_id, client_request_id, request_id, created_at FROM usage_logs ORDER BY created_at DESC LIMIT 3;" 2>&1

echo ""
echo "=== 检查迁移状态 ==="
psql -U "$PGUSER" -d "$PGDATABASE" -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 5;" 2>&1

echo ""
echo "=== 如果上面显示 client_request_id 字段不存在，请运行迁移 ==="
echo "psql -U \$PGUSER -d \$PGDATABASE -f migrations/046_add_client_request_id.sql"
