package admin

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RequestHandler handles admin request log related requests
type RequestHandler struct {
	requestLogRepo service.RequestLogRepository
	adminService   service.AdminService
	apiKeyService  *service.APIKeyService
	cleanupService *service.RequestLogCleanupService
}

// NewRequestHandler creates a new admin request handler
func NewRequestHandler(
	requestLogRepo service.RequestLogRepository,
	adminService service.AdminService,
	apiKeyService *service.APIKeyService,
	cleanupService *service.RequestLogCleanupService,
) *RequestHandler {
	return &RequestHandler{
		requestLogRepo: requestLogRepo,
		adminService:   adminService,
		apiKeyService:  apiKeyService,
		cleanupService: cleanupService,
	}
}

// CreateRequestLogCleanupTaskRequest represents cleanup task creation request
type CreateRequestLogCleanupTaskRequest struct {
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	UserID    *int64  `json:"user_id"`
	APIKeyID  *int64  `json:"api_key_id"`
	AccountID *int64  `json:"account_id"`
	GroupID   *int64  `json:"group_id"`
	Model     *string `json:"model"`
	Stream    *bool   `json:"stream"`
	IsError   *bool   `json:"is_error"`
	Timezone  string  `json:"timezone"`
}

// List handles listing all request logs with filters (admin can see all)
// GET /api/v1/admin/requests
func (h *RequestHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	// Parse filters
	var userID, apiKeyID, accountID, groupID int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = id
	}

	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		apiKeyID = id
	}

	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		id, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid account_id")
			return
		}
		accountID = id
	}

	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		id, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		groupID = id
	}

	model := c.Query("model")

	var stream *bool
	if streamStr := c.Query("stream"); streamStr != "" {
		val, err := strconv.ParseBool(streamStr)
		if err != nil {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return
		}
		stream = &val
	}

	var isError *bool
	if isErrorStr := c.Query("is_error"); isErrorStr != "" {
		val, err := strconv.ParseBool(isErrorStr)
		if err != nil {
			response.BadRequest(c, "Invalid is_error value, use true or false")
			return
		}
		isError = &val
	}

	// Parse time range
	userTZ := c.Query("timezone")
	if userTZ == "" {
		userTZ = "UTC"
	}
	var startTime, endTime *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", startDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return
		}
		startTime = &parsed
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		parsed, err := timezone.ParseInUserLocation("2006-01-02", endDateStr, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return
		}
		// Set to end of day
		endOfDay := parsed.Add(24*time.Hour - time.Second)
		endTime = &endOfDay
	}

	// Build filters (admin can query all, no userID restriction)
	filters := service.RequestLogFilters{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		AccountID: accountID,
		GroupID:   groupID,
		Model:     model,
		Stream:    stream,
		IsError:   isError,
		StartTime: startTime,
		EndTime:   endTime,
	}

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	// Query request logs
	logs, result, err := h.requestLogRepo.List(c.Request.Context(), filters, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert to admin DTO (includes IP address and account info)
	dtoLogs := make([]*dto.AdminRequestLog, len(logs))
	for i, log := range logs {
		dtoLogs[i] = dto.RequestLogFromServiceAdmin(log)
	}

	response.Paginated(c, dtoLogs, result.Total, page, pageSize)
}

// GetByID handles getting a single request log by ID (admin can see all)
// GET /api/v1/admin/requests/:id
func (h *RequestHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid id")
		return
	}

	// Get request log
	log, err := h.requestLogRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.RequestLogFromServiceAdmin(log))
}

// SearchUsers searches users for filtering
// GET /api/v1/admin/requests/search-users
func (h *RequestHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Success(c, []any{})
		return
	}

	// Limit to 30 results
	users, _, err := h.adminService.ListUsers(c.Request.Context(), 1, 30, service.UserListFilters{Search: keyword})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert to simple DTO
	type UserResult struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}

	results := make([]UserResult, len(users))
	for i, user := range users {
		results[i] = UserResult{
			ID:    user.ID,
			Email: user.Email,
		}
	}

	response.Success(c, results)
}

// SearchAPIKeys searches API keys for filtering
// GET /api/v1/admin/requests/search-api-keys
func (h *RequestHandler) SearchAPIKeys(c *gin.Context) {
	var userID int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = id
	}

	keyword := c.Query("q")

	apiKeys, err := h.apiKeyService.SearchAPIKeys(c.Request.Context(), userID, keyword, 30)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Convert to simple DTO
	type APIKeyResult struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Key  string `json:"key"`
	}

	results := make([]APIKeyResult, len(apiKeys))
	for i, apiKey := range apiKeys {
		results[i] = APIKeyResult{
			ID:   apiKey.ID,
			Name: apiKey.Name,
			Key:  apiKey.Key,
		}
	}

	response.Success(c, results)
}

// ListCleanupTasks handles listing request log cleanup tasks
// GET /api/v1/admin/requests/cleanup-tasks
func (h *RequestHandler) ListCleanupTasks(c *gin.Context) {
	if h.cleanupService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Request log cleanup service unavailable")
		return
	}
	operator := int64(0)
	if subject, ok := middleware.GetAuthSubjectFromContext(c); ok {
		operator = subject.UserID
	}
	page, pageSize := response.ParsePagination(c)
	log.Printf("[RequestLogCleanup] 请求清理任务列表: operator=%d page=%d page_size=%d", operator, page, pageSize)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	tasks, result, err := h.cleanupService.ListTasks(c.Request.Context(), params)
	if err != nil {
		log.Printf("[RequestLogCleanup] 查询清理任务列表失败: operator=%d page=%d page_size=%d err=%v", operator, page, pageSize, err)
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.RequestLogCleanupTask, 0, len(tasks))
	for i := range tasks {
		out = append(out, *dto.RequestLogCleanupTaskFromService(&tasks[i]))
	}
	log.Printf("[RequestLogCleanup] 返回清理任务列表: operator=%d total=%d items=%d page=%d page_size=%d", operator, result.Total, len(out), page, pageSize)
	response.Paginated(c, out, result.Total, page, pageSize)
}

// CreateCleanupTask handles creating a request log cleanup task
// POST /api/v1/admin/requests/cleanup-tasks
func (h *RequestHandler) CreateCleanupTask(c *gin.Context) {
	if h.cleanupService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Request log cleanup service unavailable")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req CreateRequestLogCleanupTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	req.StartDate = strings.TrimSpace(req.StartDate)
	req.EndDate = strings.TrimSpace(req.EndDate)
	if req.StartDate == "" || req.EndDate == "" {
		response.BadRequest(c, "start_date and end_date are required")
		return
	}

	startTime, err := timezone.ParseInUserLocation("2006-01-02", req.StartDate, req.Timezone)
	if err != nil {
		response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
		return
	}
	endTime, err := timezone.ParseInUserLocation("2006-01-02", req.EndDate, req.Timezone)
	if err != nil {
		response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
		return
	}
	endTime = endTime.Add(24*time.Hour - time.Nanosecond)

	filters := service.RequestLogCleanupFilters{
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    req.UserID,
		APIKeyID:  req.APIKeyID,
		AccountID: req.AccountID,
		GroupID:   req.GroupID,
		Model:     req.Model,
		Stream:    req.Stream,
		IsError:   req.IsError,
	}

	log.Printf("[RequestLogCleanup] 请求创建清理任务: operator=%d start=%s end=%s tz=%q",
		subject.UserID,
		filters.StartTime.Format(time.RFC3339),
		filters.EndTime.Format(time.RFC3339),
		req.Timezone,
	)

	task, err := h.cleanupService.CreateTask(c.Request.Context(), filters, subject.UserID)
	if err != nil {
		log.Printf("[RequestLogCleanup] 创建清理任务失败: operator=%d err=%v", subject.UserID, err)
		response.ErrorFrom(c, err)
		return
	}

	log.Printf("[RequestLogCleanup] 清理任务已创建: task=%d operator=%d status=%s", task.ID, subject.UserID, task.Status)
	response.Success(c, dto.RequestLogCleanupTaskFromService(task))
}

// CancelCleanupTask handles canceling a request log cleanup task
// POST /api/v1/admin/requests/cleanup-tasks/:id/cancel
func (h *RequestHandler) CancelCleanupTask(c *gin.Context) {
	if h.cleanupService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Request log cleanup service unavailable")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	idStr := strings.TrimSpace(c.Param("id"))
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || taskID <= 0 {
		response.BadRequest(c, "Invalid task id")
		return
	}
	log.Printf("[RequestLogCleanup] 请求取消清理任务: task=%d operator=%d", taskID, subject.UserID)
	if err := h.cleanupService.CancelTask(c.Request.Context(), taskID, subject.UserID); err != nil {
		log.Printf("[RequestLogCleanup] 取消清理任务失败: task=%d operator=%d err=%v", taskID, subject.UserID, err)
		response.ErrorFrom(c, err)
		return
	}
	log.Printf("[RequestLogCleanup] 清理任务已取消: task=%d operator=%d", taskID, subject.UserID)
	response.Success(c, gin.H{"id": taskID, "status": service.RequestLogCleanupStatusCanceled})
}
