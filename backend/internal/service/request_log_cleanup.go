package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	RequestLogCleanupStatusPending   = "pending"
	RequestLogCleanupStatusRunning   = "running"
	RequestLogCleanupStatusSucceeded = "succeeded"
	RequestLogCleanupStatusFailed    = "failed"
	RequestLogCleanupStatusCanceled  = "canceled"
)

// RequestLogCleanupFilters 定义请求记录清理任务过滤条件
type RequestLogCleanupFilters struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UserID    *int64    `json:"user_id,omitempty"`
	APIKeyID  *int64    `json:"api_key_id,omitempty"`
	AccountID *int64    `json:"account_id,omitempty"`
	GroupID   *int64    `json:"group_id,omitempty"`
	Model     *string   `json:"model,omitempty"`
	Stream    *bool     `json:"stream,omitempty"`
	IsError   *bool     `json:"is_error,omitempty"`
}

// RequestLogCleanupTask 表示请求记录清理任务
type RequestLogCleanupTask struct {
	ID          int64
	Status      string
	Filters     RequestLogCleanupFilters
	CreatedBy   int64
	DeletedRows int64
	ErrorMsg    *string
	CanceledBy  *int64
	CanceledAt  *time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RequestLogCleanupRepository 定义请求记录清理任务持久层接口
type RequestLogCleanupRepository interface {
	CreateTask(ctx context.Context, task *RequestLogCleanupTask) error
	ListTasks(ctx context.Context, params pagination.PaginationParams) ([]RequestLogCleanupTask, *pagination.PaginationResult, error)
	ClaimNextPendingTask(ctx context.Context, staleRunningAfterSeconds int64) (*RequestLogCleanupTask, error)
	GetTaskStatus(ctx context.Context, taskID int64) (string, error)
	UpdateTaskProgress(ctx context.Context, taskID int64, deletedRows int64) error
	CancelTask(ctx context.Context, taskID int64, canceledBy int64) (bool, error)
	MarkTaskSucceeded(ctx context.Context, taskID int64, deletedRows int64) error
	MarkTaskFailed(ctx context.Context, taskID int64, deletedRows int64, errorMsg string) error
	DeleteRequestLogsBatch(ctx context.Context, filters RequestLogCleanupFilters, limit int) (int64, error)
}
