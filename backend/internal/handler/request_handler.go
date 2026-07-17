package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RequestHandler handles request log related requests
type RequestHandler struct {
	requestLogRepo service.RequestLogRepository
	apiKeyService  *service.APIKeyService
}

// NewRequestHandler creates a new RequestHandler
func NewRequestHandler(requestLogRepo service.RequestLogRepository, apiKeyService *service.APIKeyService) *RequestHandler {
	return &RequestHandler{
		requestLogRepo: requestLogRepo,
		apiKeyService:  apiKeyService,
	}
}

// List handles listing request logs with pagination
// GET /api/v1/requests
func (h *RequestHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	var apiKeyID int64
	if apiKeyIDStr := c.Query("api_key_id"); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}

		// Verify API Key ownership to prevent horizontal privilege escalation
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's request logs")
			return
		}

		apiKeyID = id
	}

	// Parse additional filters
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

	// Build filters (always filter by user_id for security)
	filters := service.RequestLogFilters{
		UserID:    subject.UserID,
		APIKeyID:  apiKeyID,
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

	// Convert to DTO
	dtoLogs := make([]*dto.RequestLog, len(logs))
	for i, log := range logs {
		dtoLogs[i] = dto.RequestLogFromService(log)
	}

	response.Paginated(c, dtoLogs, result.Total, page, pageSize)
}

// GetByID handles getting a single request log by ID
// GET /api/v1/requests/:id
func (h *RequestHandler) GetByID(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

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

	// Verify ownership
	if log.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this request log")
		return
	}

	response.Success(c, dto.RequestLogFromService(log))
}
