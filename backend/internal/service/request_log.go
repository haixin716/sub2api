package service

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrRequestLogNotFound = infraerrors.NotFound("REQUEST_LOG_NOT_FOUND", "request log not found")
)

// RequestLog 定义请求日志的 Service 层数据模型。
// 记录每次 API 调用的完整请求和响应内容，用于调试、审计和问题排查。
type RequestLog struct {
	ID        int64
	UserID    int64
	APIKeyID  int64
	AccountID int64
	RequestID string
	Model     string

	GroupID *int64

	// 请求信息
	RequestBody   string
	RequestMethod string
	RequestPath   string

	// 响应信息
	ResponseBody   *string
	ResponseStatus int

	// 元数据
	Stream     bool
	DurationMs *int
	IPAddress  *string
	UserAgent  *string

	// 错误信息
	IsError      bool
	ErrorMessage *string
	ErrorType    *string

	CreatedAt time.Time

	// 关联实体
	User    *User
	APIKey  *APIKey
	Account *Account
	Group   *Group
}

// IsSuccess 返回请求是否成功（状态码 2xx）
func (r *RequestLog) IsSuccess() bool {
	return r.ResponseStatus >= 200 && r.ResponseStatus < 300
}

// HasError 返回请求是否包含错误
func (r *RequestLog) HasError() bool {
	return r.IsError || !r.IsSuccess()
}

// RecordRequestInput 定义记录请求的输入参数
type RecordRequestInput struct {
	Result    *ForwardResult // Forward 结果
	APIKey    *APIKey        // API Key
	User      *User          // 用户
	Account   *Account       // 使用的账号
	UserAgent string         // User-Agent
	IPAddress string         // IP 地址

	// 捕获的请求数据
	RequestBody   string // 请求体
	RequestMethod string // HTTP 方法
	RequestPath   string // 请求路径

	// 捕获的响应数据
	ResponseBody   *string // 响应体（可能为空）
	ResponseStatus int     // HTTP 状态码
}

// CapturedRequestData 存储捕获的请求数据
type CapturedRequestData struct {
	Body   string
	Method string
	Path   string
}

// CapturedResponseData 存储捕获的响应数据
type CapturedResponseData struct {
	Body       *string
	Status     int
	IsError    bool
	ErrorMsg   *string
	ErrorType  *string
}
