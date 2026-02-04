-- 创建请求日志表
-- 记录每次 API 调用的完整请求和响应内容，用于调试、审计和问题排查

CREATE TABLE IF NOT EXISTS request_logs (
    id                  BIGSERIAL PRIMARY KEY,

    -- 关联字段
    user_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id          BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    account_id          BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    request_id          VARCHAR(64) NOT NULL,                   -- 请求唯一ID，与usage_logs表关联
    model               VARCHAR(100) NOT NULL,                  -- 模型名称
    group_id            BIGINT REFERENCES groups(id) ON DELETE SET NULL,

    -- 请求信息
    request_body        TEXT NOT NULL,                          -- 完整请求体JSON
    request_method      VARCHAR(10) NOT NULL DEFAULT 'POST',    -- HTTP方法
    request_path        VARCHAR(255) NOT NULL DEFAULT '/v1/messages', -- 请求路径

    -- 响应信息
    response_body       TEXT,                                   -- 完整响应体JSON（非流式）或聚合内容（流式）
    response_status     INT NOT NULL DEFAULT 200,               -- HTTP状态码

    -- 元数据
    stream              BOOLEAN NOT NULL DEFAULT FALSE,         -- 是否流式响应
    duration_ms         INT,                                    -- 请求耗时（毫秒）
    ip_address          VARCHAR(45),                            -- 客户端IP地址（支持IPv6）
    user_agent          VARCHAR(512),                           -- User-Agent头信息

    -- 错误信息
    is_error            BOOLEAN NOT NULL DEFAULT FALSE,         -- 是否错误请求
    error_message       TEXT,                                   -- 错误信息
    error_type          VARCHAR(50),                            -- 错误类型

    -- 时间戳（只追加表，不可修改）
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()      -- 创建时间
);

-- 创建索引，优化查询性能
CREATE INDEX IF NOT EXISTS idx_request_logs_user_id ON request_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_api_key_id ON request_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_account_id ON request_logs(account_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_group_id ON request_logs(group_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_request_id ON request_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model);
CREATE INDEX IF NOT EXISTS idx_request_logs_is_error ON request_logs(is_error);

-- 复合索引，用于时间范围查询
CREATE INDEX IF NOT EXISTS idx_request_logs_user_created ON request_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_api_key_created ON request_logs(api_key_id, created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_error_created ON request_logs(is_error, created_at);

-- 添加表注释
COMMENT ON TABLE request_logs IS '请求日志表，记录每次API调用的完整请求和响应内容';
COMMENT ON COLUMN request_logs.request_id IS '请求唯一ID，与usage_logs表关联';
COMMENT ON COLUMN request_logs.request_body IS '完整请求体JSON';
COMMENT ON COLUMN request_logs.response_body IS '完整响应体JSON（非流式）或聚合内容（流式）';
COMMENT ON COLUMN request_logs.is_error IS '是否错误请求';
COMMENT ON COLUMN request_logs.stream IS '是否流式响应';
