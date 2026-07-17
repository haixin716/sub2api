package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// RequestLog is the user-visible request log payload. Administrator-only
// fields such as the client IP and upstream account are intentionally omitted.
type RequestLog struct {
	ID              int64   `json:"id"`
	UserID          int64   `json:"user_id"`
	APIKeyID        int64   `json:"api_key_id"`
	AccountID       int64   `json:"account_id"`
	ClientRequestID string  `json:"client_request_id"`
	RequestID       *string `json:"request_id,omitempty"`
	Model           string  `json:"model"`
	GroupID         *int64  `json:"group_id"`

	RequestBody   string `json:"request_body"`
	RequestMethod string `json:"request_method"`
	RequestPath   string `json:"request_path"`

	ResponseBody   *string `json:"response_body"`
	ResponseStatus int     `json:"response_status"`
	Stream         bool    `json:"stream"`
	DurationMs     *int    `json:"duration_ms"`
	UserAgent      *string `json:"user_agent,omitempty"`

	IsError      bool      `json:"is_error"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	ErrorType    *string   `json:"error_type,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	User   *User   `json:"user,omitempty"`
	APIKey *APIKey `json:"api_key,omitempty"`
	Group  *Group  `json:"group,omitempty"`
}

// AdminRequestLog adds administrator-only request metadata.
type AdminRequestLog struct {
	RequestLog
	IPAddress *string         `json:"ip_address,omitempty"`
	Account   *AccountSummary `json:"account,omitempty"`
}

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

type RequestLogCleanupTask struct {
	ID           int64                    `json:"id"`
	Status       string                   `json:"status"`
	Filters      RequestLogCleanupFilters `json:"filters"`
	CreatedBy    int64                    `json:"created_by"`
	DeletedRows  int64                    `json:"deleted_rows"`
	ErrorMessage *string                  `json:"error_message,omitempty"`
	CanceledBy   *int64                   `json:"canceled_by,omitempty"`
	CanceledAt   *time.Time               `json:"canceled_at,omitempty"`
	StartedAt    *time.Time               `json:"started_at,omitempty"`
	FinishedAt   *time.Time               `json:"finished_at,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

func requestLogFromServiceUser(log *service.RequestLog) RequestLog {
	return RequestLog{
		ID:              log.ID,
		UserID:          log.UserID,
		APIKeyID:        log.APIKeyID,
		AccountID:       log.AccountID,
		ClientRequestID: log.ClientRequestID,
		RequestID:       log.RequestID,
		Model:           log.Model,
		GroupID:         log.GroupID,
		RequestBody:     log.RequestBody,
		RequestMethod:   log.RequestMethod,
		RequestPath:     log.RequestPath,
		ResponseBody:    log.ResponseBody,
		ResponseStatus:  log.ResponseStatus,
		Stream:          log.Stream,
		DurationMs:      log.DurationMs,
		UserAgent:       log.UserAgent,
		IsError:         log.IsError,
		ErrorMessage:    log.ErrorMessage,
		ErrorType:       log.ErrorType,
		CreatedAt:       log.CreatedAt,
		User:            UserFromServiceShallow(log.User),
		APIKey:          APIKeyFromService(log.APIKey),
		Group:           GroupFromServiceShallow(log.Group),
	}
}

func RequestLogFromService(log *service.RequestLog) *RequestLog {
	if log == nil {
		return nil
	}
	out := requestLogFromServiceUser(log)
	return &out
}

func RequestLogFromServiceAdmin(log *service.RequestLog) *AdminRequestLog {
	if log == nil {
		return nil
	}
	return &AdminRequestLog{
		RequestLog: requestLogFromServiceUser(log),
		IPAddress:  log.IPAddress,
		Account:    AccountSummaryFromService(log.Account),
	}
}

func RequestLogCleanupTaskFromService(task *service.RequestLogCleanupTask) *RequestLogCleanupTask {
	if task == nil {
		return nil
	}
	return &RequestLogCleanupTask{
		ID:     task.ID,
		Status: task.Status,
		Filters: RequestLogCleanupFilters{
			StartTime: task.Filters.StartTime,
			EndTime:   task.Filters.EndTime,
			UserID:    task.Filters.UserID,
			APIKeyID:  task.Filters.APIKeyID,
			AccountID: task.Filters.AccountID,
			GroupID:   task.Filters.GroupID,
			Model:     task.Filters.Model,
			Stream:    task.Filters.Stream,
			IsError:   task.Filters.IsError,
		},
		CreatedBy:    task.CreatedBy,
		DeletedRows:  task.DeletedRows,
		ErrorMessage: task.ErrorMsg,
		CanceledBy:   task.CanceledBy,
		CanceledAt:   task.CanceledAt,
		StartedAt:    task.StartedAt,
		FinishedAt:   task.FinishedAt,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}
