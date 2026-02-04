package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// RequestLogRepository 定义请求日志的数据访问接口
type RequestLogRepository interface {
	// Create 创建请求日志记录
	Create(ctx context.Context, log *RequestLog) error

	// GetByID 根据ID获取请求日志
	GetByID(ctx context.Context, id int64) (*RequestLog, error)

	// GetByRequestID 根据请求ID获取请求日志
	GetByRequestID(ctx context.Context, requestID string) (*RequestLog, error)

	// List 查询请求日志列表（支持分页和过滤）
	List(ctx context.Context, filters RequestLogFilters, params pagination.PaginationParams) ([]*RequestLog, *pagination.PaginationResult, error)

	// Count 统计请求日志数量
	Count(ctx context.Context, filters RequestLogFilters) (int64, error)
}

// RequestLogFilters 定义请求日志的过滤条件
type RequestLogFilters struct {
	UserID    int64  // 用户ID（0表示不过滤）
	APIKeyID  int64  // API Key ID（0表示不过滤）
	AccountID int64  // 账号ID（0表示不过滤）
	GroupID   int64  // 分组ID（0表示不过滤）
	Model     string // 模型名称（空表示不过滤）

	Stream      *bool // 是否流式（nil表示不过滤）
	IsError     *bool // 是否错误（nil表示不过滤）
	StatusCode  *int  // HTTP状态码（nil表示不过滤）
	RequestPath string // 请求路径（空表示不过滤）

	StartTime *time.Time // 开始时间（nil表示不限制）
	EndTime   *time.Time // 结束时间（nil表示不限制）
}
