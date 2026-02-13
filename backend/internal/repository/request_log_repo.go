package repository

import (
	"context"
	"database/sql"
	"fmt"
	stdlog "log"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbrequestlog "github.com/Wei-Shaw/sub2api/ent/requestlog"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type requestLogRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewRequestLogRepository 创建请求日志 Repository
func NewRequestLogRepository(client *dbent.Client, sqlDB *sql.DB) service.RequestLogRepository {
	return &requestLogRepository{
		client: client,
		sql:    sqlDB,
	}
}

// Create 创建请求日志记录
func (r *requestLogRepository) Create(ctx context.Context, log *service.RequestLog) error {
	if log == nil {
		return nil
	}

	// 验证 client_request_id 不能为空
	if strings.TrimSpace(log.ClientRequestID) == "" {
		stdlog.Printf("[ERROR] RequestLogRepo.Create: client_request_id is empty! UserID=%d, APIKeyID=%d", log.UserID, log.APIKeyID)
		return fmt.Errorf("client_request_id is required")
	}

	stdlog.Printf("[DEBUG] RequestLogRepo.Create: Inserting with ClientRequestID='%s', UserID=%d, APIKeyID=%d, AccountID=%d",
		log.ClientRequestID, log.UserID, log.APIKeyID, log.AccountID)

	// 在事务上下文中，使用 tx 绑定的 ExecQuerier 执行原生 SQL
	sqlq := r.sql
	if tx := dbent.TxFromContext(ctx); tx != nil {
		sqlq = tx.Client()
	}

	createdAt := log.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	query := `
		INSERT INTO request_logs (
			user_id,
			api_key_id,
			account_id,
			client_request_id,
			request_id,
			model,
			group_id,
			request_body,
			request_method,
			request_path,
			response_body,
			response_status,
			stream,
			duration_ms,
			ip_address,
			user_agent,
			is_error,
			error_message,
			error_type,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (client_request_id, api_key_id) DO NOTHING
	`

	args := []any{
		log.UserID,
		log.APIKeyID,
		log.AccountID,
		log.ClientRequestID,
		log.RequestID,
		log.Model,
		log.GroupID,
		log.RequestBody,
		log.RequestMethod,
		log.RequestPath,
		log.ResponseBody,
		log.ResponseStatus,
		log.Stream,
		log.DurationMs,
		log.IPAddress,
		log.UserAgent,
		log.IsError,
		log.ErrorMessage,
		log.ErrorType,
		createdAt,
	}

	_, err := sqlq.ExecContext(ctx, query, args...)
	if err != nil {
		stdlog.Printf("[ERROR] RequestLogRepo.Create: Failed to insert request log: %v", err)
		return fmt.Errorf("failed to create request log: %w", err)
	}

	stdlog.Printf("[DEBUG] RequestLogRepo.Create: Successfully inserted request log with ClientRequestID='%s'", log.ClientRequestID)
	return nil
}

// GetByID 根据ID获取请求日志
func (r *requestLogRepository) GetByID(ctx context.Context, id int64) (*service.RequestLog, error) {
	dbLog, err := r.client.RequestLog.Query().
		Where(dbrequestlog.ID(id)).
		WithUser().
		WithAPIKey().
		WithAccount().
		WithGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRequestLogNotFound
		}
		return nil, fmt.Errorf("failed to get request log: %w", err)
	}

	return mapRequestLogFromEnt(dbLog), nil
}

// GetByRequestID 根据请求ID获取请求日志
func (r *requestLogRepository) GetByRequestID(ctx context.Context, requestID string) (*service.RequestLog, error) {
	dbLog, err := r.client.RequestLog.Query().
		Where(dbrequestlog.RequestID(requestID)).
		WithUser().
		WithAPIKey().
		WithAccount().
		WithGroup().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrRequestLogNotFound
		}
		return nil, fmt.Errorf("failed to get request log by request_id: %w", err)
	}

	return mapRequestLogFromEnt(dbLog), nil
}

// List 查询请求日志列表（支持分页和过滤）
func (r *requestLogRepository) List(ctx context.Context, filters service.RequestLogFilters, params pagination.PaginationParams) ([]*service.RequestLog, *pagination.PaginationResult, error) {
	query := r.client.RequestLog.Query()

	// 应用过滤条件
	query = applyRequestLogFilters(query, filters)

	// 计算总数
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count request logs: %w", err)
	}

	// 应用分页
	offset := (params.Page - 1) * params.PageSize
	query = query.
		Limit(params.PageSize).
		Offset(offset).
		Order(dbent.Desc(dbrequestlog.FieldCreatedAt))

	// 加载关联数据
	query = query.
		WithUser().
		WithAPIKey().
		WithAccount().
		WithGroup()

	// 查询数据
	dbLogs, err := query.All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list request logs: %w", err)
	}

	// 转换为 service 模型
	logs := make([]*service.RequestLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = mapRequestLogFromEnt(dbLog)
	}

	// 构建分页结果
	result := &pagination.PaginationResult{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(total),
		Pages:    (total + params.PageSize - 1) / params.PageSize,
	}

	return logs, result, nil
}

// Count 统计请求日志数量
func (r *requestLogRepository) Count(ctx context.Context, filters service.RequestLogFilters) (int64, error) {
	query := r.client.RequestLog.Query()
	query = applyRequestLogFilters(query, filters)

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count request logs: %w", err)
	}

	return int64(count), nil
}

// applyRequestLogFilters 应用过滤条件
func applyRequestLogFilters(query *dbent.RequestLogQuery, filters service.RequestLogFilters) *dbent.RequestLogQuery {
	if filters.UserID > 0 {
		query = query.Where(dbrequestlog.UserID(filters.UserID))
	}
	if filters.APIKeyID > 0 {
		query = query.Where(dbrequestlog.APIKeyID(filters.APIKeyID))
	}
	if filters.AccountID > 0 {
		query = query.Where(dbrequestlog.AccountID(filters.AccountID))
	}
	if filters.GroupID > 0 {
		query = query.Where(dbrequestlog.GroupID(filters.GroupID))
	}
	if filters.Model != "" {
		query = query.Where(dbrequestlog.Model(filters.Model))
	}
	if filters.Stream != nil {
		query = query.Where(dbrequestlog.Stream(*filters.Stream))
	}
	if filters.IsError != nil {
		query = query.Where(dbrequestlog.IsError(*filters.IsError))
	}
	if filters.StatusCode != nil {
		query = query.Where(dbrequestlog.ResponseStatus(*filters.StatusCode))
	}
	if filters.RequestPath != "" {
		query = query.Where(dbrequestlog.RequestPathContains(filters.RequestPath))
	}
	if filters.StartTime != nil {
		query = query.Where(dbrequestlog.CreatedAtGTE(*filters.StartTime))
	}
	if filters.EndTime != nil {
		query = query.Where(dbrequestlog.CreatedAtLTE(*filters.EndTime))
	}
	return query
}

// mapRequestLogFromEnt 将 Ent 模型转换为 Service 模型
func mapRequestLogFromEnt(dbLog *dbent.RequestLog) *service.RequestLog {
	if dbLog == nil {
		return nil
	}

	log := &service.RequestLog{
		ID:              dbLog.ID,
		UserID:          dbLog.UserID,
		APIKeyID:        dbLog.APIKeyID,
		AccountID:       dbLog.AccountID,
		ClientRequestID: dbLog.ClientRequestID,
		RequestID:       dbLog.RequestID,
		Model:           dbLog.Model,
		GroupID:         dbLog.GroupID,
		RequestBody:     dbLog.RequestBody,
		RequestMethod:   dbLog.RequestMethod,
		RequestPath:     dbLog.RequestPath,
		ResponseBody:    dbLog.ResponseBody,
		ResponseStatus:  dbLog.ResponseStatus,
		Stream:          dbLog.Stream,
		DurationMs:      dbLog.DurationMs,
		IPAddress:       dbLog.IPAddress,
		UserAgent:       dbLog.UserAgent,
		IsError:         dbLog.IsError,
		ErrorMessage:    dbLog.ErrorMessage,
		ErrorType:       dbLog.ErrorType,
		CreatedAt:       dbLog.CreatedAt,
	}

	// 映射关联实体
	if dbLog.Edges.User != nil {
		log.User = &service.User{
			ID:    dbLog.Edges.User.ID,
			Email: dbLog.Edges.User.Email,
		}
	}
	if dbLog.Edges.APIKey != nil {
		log.APIKey = &service.APIKey{
			ID:   dbLog.Edges.APIKey.ID,
			Name: dbLog.Edges.APIKey.Name,
			Key:  dbLog.Edges.APIKey.Key,
		}
	}
	if dbLog.Edges.Account != nil {
		log.Account = &service.Account{
			ID:   dbLog.Edges.Account.ID,
			Name: dbLog.Edges.Account.Name,
		}
	}
	if dbLog.Edges.Group != nil {
		log.Group = &service.Group{
			ID:   dbLog.Edges.Group.ID,
			Name: dbLog.Edges.Group.Name,
		}
	}

	return log
}
