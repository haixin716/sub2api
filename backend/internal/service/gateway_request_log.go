package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/google/uuid"
)

// RecordRequest persists the request/response metadata captured by the gateway
// after a successful upstream forward. The stable client request ID comes from
// the current f_dev_3.0 request context; the upstream request ID remains a
// separate optional field.
func (s *GatewayService) RecordRequest(ctx context.Context, input *RecordRequestInput) error {
	if s == nil || s.requestLogRepo == nil || input == nil || input.Result == nil {
		return nil
	}
	if input.APIKey == nil || input.User == nil || input.Account == nil {
		return fmt.Errorf("record request: missing api key, user, or account")
	}

	result := input.Result
	clientRequestID := ""
	if ctx != nil {
		clientRequestID, _ = ctx.Value(ctxkey.ClientRequestID).(string)
		clientRequestID = strings.TrimSpace(clientRequestID)
	}
	if clientRequestID == "" {
		clientRequestID = strings.TrimSpace(result.RequestID)
	}
	if clientRequestID == "" {
		clientRequestID = uuid.NewString()
	}

	var upstreamRequestID *string
	if requestID := strings.TrimSpace(result.RequestID); requestID != "" {
		upstreamRequestID = &requestID
	}

	status := input.ResponseStatus
	if status == 0 {
		status = 200
	}
	isError := status < 200 || status >= 400
	errorMessage := input.ResponseBody
	errorType := requestLogErrorType(status, isError)
	durationMs := int(result.Duration.Milliseconds())

	logEntry := &RequestLog{
		UserID:          input.User.ID,
		APIKeyID:        input.APIKey.ID,
		AccountID:       input.Account.ID,
		ClientRequestID: clientRequestID,
		RequestID:       upstreamRequestID,
		Model:           result.Model,
		GroupID:         input.APIKey.GroupID,
		RequestBody:     input.RequestBody,
		RequestMethod:   input.RequestMethod,
		RequestPath:     input.RequestPath,
		ResponseBody:    input.ResponseBody,
		ResponseStatus:  status,
		Stream:          result.Stream,
		DurationMs:      &durationMs,
		IsError:         isError,
		ErrorMessage:    errorMessage,
		ErrorType:       errorType,
	}
	if value := strings.TrimSpace(input.IPAddress); value != "" {
		logEntry.IPAddress = &value
	}
	if value := strings.TrimSpace(input.UserAgent); value != "" {
		logEntry.UserAgent = &value
	}

	if err := s.requestLogRepo.Create(ctx, logEntry); err != nil {
		return fmt.Errorf("record request: %w", err)
	}
	return nil
}

func requestLogErrorType(status int, isError bool) *string {
	if !isError {
		return nil
	}
	value := "unknown_error"
	switch status {
	case 400:
		value = "invalid_request_error"
	case 401, 403:
		value = "authentication_error"
	case 429:
		value = "rate_limit_error"
	case 500, 502, 503, 504:
		value = "api_error"
	}
	return &value
}
