package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RequestHandler handles admin request log related requests
type RequestHandler struct {
	requestLogRepo service.RequestLogRepository
	adminService   service.AdminService
	apiKeyService  *service.APIKeyService
}

// NewRequestHandler creates a new admin request handler
func NewRequestHandler(
	requestLogRepo service.RequestLogRepository,
	adminService service.AdminService,
	apiKeyService *service.APIKeyService,
) *RequestHandler {
	return &RequestHandler{
		requestLogRepo: requestLogRepo,
		adminService:   adminService,
		apiKeyService:  apiKeyService,
	}
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
