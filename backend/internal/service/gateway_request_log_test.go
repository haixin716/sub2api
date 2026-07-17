package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type requestLogRepositoryStub struct {
	created *RequestLog
}

func (s *requestLogRepositoryStub) Create(_ context.Context, log *RequestLog) error {
	s.created = log
	return nil
}

func (s *requestLogRepositoryStub) GetByID(context.Context, int64) (*RequestLog, error) {
	return nil, ErrRequestLogNotFound
}

func (s *requestLogRepositoryStub) GetByRequestID(context.Context, string) (*RequestLog, error) {
	return nil, ErrRequestLogNotFound
}

func (s *requestLogRepositoryStub) List(context.Context, RequestLogFilters, pagination.PaginationParams) ([]*RequestLog, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *requestLogRepositoryStub) Count(context.Context, RequestLogFilters) (int64, error) {
	return 0, nil
}

func TestGatewayServiceRecordRequestUsesCurrentClientRequestID(t *testing.T) {
	repo := &requestLogRepositoryStub{}
	svc := &GatewayService{requestLogRepo: repo}
	groupID := int64(9)
	responseBody := `{"type":"error","error":{"message":"limited"}}`
	ctx := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-request-123")

	err := svc.RecordRequest(ctx, &RecordRequestInput{
		Result: &ForwardResult{
			RequestID: "upstream-request-456",
			Model:     "claude-test",
			Stream:    true,
			Duration:  1250 * time.Millisecond,
		},
		APIKey:         &APIKey{ID: 2, GroupID: &groupID},
		User:           &User{ID: 1},
		Account:        &Account{ID: 3},
		UserAgent:      "test-agent",
		IPAddress:      "127.0.0.1",
		RequestBody:    `{"model":"claude-test"}`,
		RequestMethod:  "POST",
		RequestPath:    "/v1/messages",
		ResponseBody:   &responseBody,
		ResponseStatus: 429,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.created)
	require.Equal(t, "client-request-123", repo.created.ClientRequestID)
	require.Equal(t, "upstream-request-456", requireStringPointer(t, repo.created.RequestID))
	require.Equal(t, groupID, requireInt64Pointer(t, repo.created.GroupID))
	require.Equal(t, 1250, requireIntPointer(t, repo.created.DurationMs))
	require.True(t, repo.created.IsError)
	require.Equal(t, "rate_limit_error", requireStringPointer(t, repo.created.ErrorType))
	require.Equal(t, responseBody, requireStringPointer(t, repo.created.ErrorMessage))
}

func requireStringPointer(t *testing.T, value *string) string {
	t.Helper()
	require.NotNil(t, value)
	return *value
}

func requireInt64Pointer(t *testing.T, value *int64) int64 {
	t.Helper()
	require.NotNil(t, value)
	return *value
}

func requireIntPointer(t *testing.T, value *int) int {
	t.Helper()
	require.NotNil(t, value)
	return *value
}
