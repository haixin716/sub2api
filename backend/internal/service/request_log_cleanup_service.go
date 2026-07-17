package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	requestLogCleanupWorkerName = "request_log_cleanup_worker"
)

// RequestLogCleanupService 负责创建与执行请求记录清理任务
type RequestLogCleanupService struct {
	repo        RequestLogCleanupRepository
	timingWheel *TimingWheelService
	cfg         *config.Config

	running   int32
	startOnce sync.Once
	stopOnce  sync.Once

	workerCtx    context.Context
	workerCancel context.CancelFunc
}

func NewRequestLogCleanupService(repo RequestLogCleanupRepository, timingWheel *TimingWheelService, cfg *config.Config) *RequestLogCleanupService {
	workerCtx, workerCancel := context.WithCancel(context.Background())
	return &RequestLogCleanupService{
		repo:         repo,
		timingWheel:  timingWheel,
		cfg:          cfg,
		workerCtx:    workerCtx,
		workerCancel: workerCancel,
	}
}

func describeRequestLogCleanupFilters(filters RequestLogCleanupFilters) string {
	var parts []string
	parts = append(parts, "start="+filters.StartTime.UTC().Format(time.RFC3339))
	parts = append(parts, "end="+filters.EndTime.UTC().Format(time.RFC3339))
	if filters.UserID != nil {
		parts = append(parts, fmt.Sprintf("user_id=%d", *filters.UserID))
	}
	if filters.APIKeyID != nil {
		parts = append(parts, fmt.Sprintf("api_key_id=%d", *filters.APIKeyID))
	}
	if filters.AccountID != nil {
		parts = append(parts, fmt.Sprintf("account_id=%d", *filters.AccountID))
	}
	if filters.GroupID != nil {
		parts = append(parts, fmt.Sprintf("group_id=%d", *filters.GroupID))
	}
	if filters.Model != nil {
		parts = append(parts, "model="+strings.TrimSpace(*filters.Model))
	}
	if filters.Stream != nil {
		parts = append(parts, fmt.Sprintf("stream=%t", *filters.Stream))
	}
	if filters.IsError != nil {
		parts = append(parts, fmt.Sprintf("is_error=%t", *filters.IsError))
	}
	return strings.Join(parts, " ")
}

func (s *RequestLogCleanupService) Start() {
	if s == nil {
		return
	}
	if s.cfg != nil && !s.cfg.UsageCleanup.Enabled {
		log.Printf("[RequestLogCleanup] not started (disabled)")
		return
	}
	if s.repo == nil || s.timingWheel == nil {
		log.Printf("[RequestLogCleanup] not started (missing deps)")
		return
	}

	interval := s.workerInterval()
	s.startOnce.Do(func() {
		s.timingWheel.ScheduleRecurring(requestLogCleanupWorkerName, interval, s.runOnce)
		log.Printf("[RequestLogCleanup] started (interval=%s max_range_days=%d batch_size=%d task_timeout=%s)", interval, s.maxRangeDays(), s.batchSize(), s.taskTimeout())
	})
}

func (s *RequestLogCleanupService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.workerCancel != nil {
			s.workerCancel()
		}
		if s.timingWheel != nil {
			s.timingWheel.Cancel(requestLogCleanupWorkerName)
		}
		log.Printf("[RequestLogCleanup] stopped")
	})
}

func (s *RequestLogCleanupService) ListTasks(ctx context.Context, params pagination.PaginationParams) ([]RequestLogCleanupTask, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, fmt.Errorf("cleanup service not ready")
	}
	return s.repo.ListTasks(ctx, params)
}

func (s *RequestLogCleanupService) CreateTask(ctx context.Context, filters RequestLogCleanupFilters, createdBy int64) (*RequestLogCleanupTask, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("cleanup service not ready")
	}
	if s.cfg != nil && !s.cfg.UsageCleanup.Enabled {
		return nil, infraerrors.New(http.StatusServiceUnavailable, "REQUEST_LOG_CLEANUP_DISABLED", "request log cleanup is disabled")
	}
	if createdBy <= 0 {
		return nil, infraerrors.BadRequest("REQUEST_LOG_CLEANUP_INVALID_CREATOR", "invalid creator")
	}

	log.Printf("[RequestLogCleanup] create_task requested: operator=%d %s", createdBy, describeRequestLogCleanupFilters(filters))
	sanitizeRequestLogCleanupFilters(&filters)
	if err := s.validateFilters(filters); err != nil {
		log.Printf("[RequestLogCleanup] create_task rejected: operator=%d err=%v %s", createdBy, err, describeRequestLogCleanupFilters(filters))
		return nil, err
	}

	task := &RequestLogCleanupTask{
		Status:    RequestLogCleanupStatusPending,
		Filters:   filters,
		CreatedBy: createdBy,
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		log.Printf("[RequestLogCleanup] create_task persist failed: operator=%d err=%v %s", createdBy, err, describeRequestLogCleanupFilters(filters))
		return nil, fmt.Errorf("create cleanup task: %w", err)
	}
	log.Printf("[RequestLogCleanup] create_task persisted: task=%d operator=%d status=%s deleted_rows=%d %s", task.ID, createdBy, task.Status, task.DeletedRows, describeRequestLogCleanupFilters(filters))
	go s.runOnce()
	return task, nil
}

func (s *RequestLogCleanupService) runOnce() {
	svc := s
	if svc == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&svc.running, 0, 1) {
		log.Printf("[RequestLogCleanup] run_once skipped: already_running=true")
		return
	}
	defer atomic.StoreInt32(&svc.running, 0)

	parent := context.Background()
	if svc.workerCtx != nil {
		parent = svc.workerCtx
	}
	ctx, cancel := context.WithTimeout(parent, svc.taskTimeout())
	defer cancel()

	task, err := svc.repo.ClaimNextPendingTask(ctx, int64(svc.taskTimeout().Seconds()))
	if err != nil {
		log.Printf("[RequestLogCleanup] claim pending task failed: %v", err)
		return
	}
	if task == nil {
		log.Printf("[RequestLogCleanup] run_once done: no_task=true")
		return
	}

	log.Printf("[RequestLogCleanup] task claimed: task=%d status=%s created_by=%d deleted_rows=%d %s", task.ID, task.Status, task.CreatedBy, task.DeletedRows, describeRequestLogCleanupFilters(task.Filters))
	svc.executeTask(ctx, task)
}

func (s *RequestLogCleanupService) executeTask(ctx context.Context, task *RequestLogCleanupTask) {
	if task == nil {
		return
	}

	batchSize := s.batchSize()
	deletedTotal := task.DeletedRows
	start := time.Now()
	log.Printf("[RequestLogCleanup] task started: task=%d batch_size=%d deleted_rows=%d %s", task.ID, batchSize, deletedTotal, describeRequestLogCleanupFilters(task.Filters))
	var batchNum int

	for {
		if ctx != nil && ctx.Err() != nil {
			log.Printf("[RequestLogCleanup] task interrupted: task=%d err=%v", task.ID, ctx.Err())
			return
		}
		canceled, err := s.isTaskCanceled(ctx, task.ID)
		if err != nil {
			s.markTaskFailed(task.ID, deletedTotal, err)
			return
		}
		if canceled {
			log.Printf("[RequestLogCleanup] task canceled: task=%d deleted_rows=%d duration=%s", task.ID, deletedTotal, time.Since(start))
			return
		}

		batchNum++
		deleted, err := s.repo.DeleteRequestLogsBatch(ctx, task.Filters, batchSize)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Printf("[RequestLogCleanup] task interrupted: task=%d err=%v", task.ID, err)
				return
			}
			s.markTaskFailed(task.ID, deletedTotal, err)
			return
		}
		deletedTotal += deleted
		if deleted > 0 {
			updateCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := s.repo.UpdateTaskProgress(updateCtx, task.ID, deletedTotal); err != nil {
				log.Printf("[RequestLogCleanup] task progress update failed: task=%d deleted_rows=%d err=%v", task.ID, deletedTotal, err)
			}
			cancel()
		}
		if batchNum <= 3 || batchNum%20 == 0 || deleted < int64(batchSize) {
			log.Printf("[RequestLogCleanup] task batch done: task=%d batch=%d deleted=%d deleted_total=%d", task.ID, batchNum, deleted, deletedTotal)
		}
		if deleted == 0 || deleted < int64(batchSize) {
			break
		}
	}

	updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.MarkTaskSucceeded(updateCtx, task.ID, deletedTotal); err != nil {
		log.Printf("[RequestLogCleanup] update task succeeded failed: task=%d err=%v", task.ID, err)
	} else {
		log.Printf("[RequestLogCleanup] task succeeded: task=%d deleted_rows=%d duration=%s", task.ID, deletedTotal, time.Since(start))
	}
}

func (s *RequestLogCleanupService) markTaskFailed(taskID int64, deletedRows int64, err error) {
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 500 {
		msg = msg[:500]
	}
	log.Printf("[RequestLogCleanup] task failed: task=%d deleted_rows=%d err=%s", taskID, deletedRows, msg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if updateErr := s.repo.MarkTaskFailed(ctx, taskID, deletedRows, msg); updateErr != nil {
		log.Printf("[RequestLogCleanup] update task failed failed: task=%d err=%v", taskID, updateErr)
	}
}

func (s *RequestLogCleanupService) isTaskCanceled(ctx context.Context, taskID int64) (bool, error) {
	if s == nil || s.repo == nil {
		return false, fmt.Errorf("cleanup service not ready")
	}
	checkCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	status, err := s.repo.GetTaskStatus(checkCtx, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if status == RequestLogCleanupStatusCanceled {
		log.Printf("[RequestLogCleanup] task cancel detected: task=%d", taskID)
	}
	return status == RequestLogCleanupStatusCanceled, nil
}

func (s *RequestLogCleanupService) validateFilters(filters RequestLogCleanupFilters) error {
	if filters.StartTime.IsZero() || filters.EndTime.IsZero() {
		return infraerrors.BadRequest("REQUEST_LOG_CLEANUP_MISSING_RANGE", "start_date and end_date are required")
	}
	if filters.EndTime.Before(filters.StartTime) {
		return infraerrors.BadRequest("REQUEST_LOG_CLEANUP_INVALID_RANGE", "end_date must be after start_date")
	}
	maxDays := s.maxRangeDays()
	if maxDays > 0 {
		delta := filters.EndTime.Sub(filters.StartTime)
		if delta > time.Duration(maxDays)*24*time.Hour {
			return infraerrors.BadRequest("REQUEST_LOG_CLEANUP_RANGE_TOO_LARGE", fmt.Sprintf("date range exceeds %d days", maxDays))
		}
	}
	return nil
}

func (s *RequestLogCleanupService) CancelTask(ctx context.Context, taskID int64, canceledBy int64) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("cleanup service not ready")
	}
	if s.cfg != nil && !s.cfg.UsageCleanup.Enabled {
		return infraerrors.New(http.StatusServiceUnavailable, "REQUEST_LOG_CLEANUP_DISABLED", "request log cleanup is disabled")
	}
	if canceledBy <= 0 {
		return infraerrors.BadRequest("REQUEST_LOG_CLEANUP_INVALID_CANCELLER", "invalid canceller")
	}
	status, err := s.repo.GetTaskStatus(ctx, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infraerrors.New(http.StatusNotFound, "REQUEST_LOG_CLEANUP_TASK_NOT_FOUND", "cleanup task not found")
		}
		return err
	}
	log.Printf("[RequestLogCleanup] cancel_task requested: task=%d operator=%d status=%s", taskID, canceledBy, status)
	if status != RequestLogCleanupStatusPending && status != RequestLogCleanupStatusRunning {
		return infraerrors.New(http.StatusConflict, "REQUEST_LOG_CLEANUP_CANCEL_CONFLICT", "cleanup task cannot be canceled in current status")
	}
	ok, err := s.repo.CancelTask(ctx, taskID, canceledBy)
	if err != nil {
		return err
	}
	if !ok {
		return infraerrors.New(http.StatusConflict, "REQUEST_LOG_CLEANUP_CANCEL_CONFLICT", "cleanup task cannot be canceled in current status")
	}
	log.Printf("[RequestLogCleanup] cancel_task done: task=%d operator=%d", taskID, canceledBy)
	return nil
}

func sanitizeRequestLogCleanupFilters(filters *RequestLogCleanupFilters) {
	if filters == nil {
		return
	}
	if filters.UserID != nil && *filters.UserID <= 0 {
		filters.UserID = nil
	}
	if filters.APIKeyID != nil && *filters.APIKeyID <= 0 {
		filters.APIKeyID = nil
	}
	if filters.AccountID != nil && *filters.AccountID <= 0 {
		filters.AccountID = nil
	}
	if filters.GroupID != nil && *filters.GroupID <= 0 {
		filters.GroupID = nil
	}
	if filters.Model != nil {
		model := strings.TrimSpace(*filters.Model)
		if model == "" {
			filters.Model = nil
		} else {
			filters.Model = &model
		}
	}
}

func (s *RequestLogCleanupService) maxRangeDays() int {
	if s == nil || s.cfg == nil {
		return 31
	}
	if s.cfg.UsageCleanup.MaxRangeDays > 0 {
		return s.cfg.UsageCleanup.MaxRangeDays
	}
	return 31
}

func (s *RequestLogCleanupService) batchSize() int {
	if s == nil || s.cfg == nil {
		return 5000
	}
	if s.cfg.UsageCleanup.BatchSize > 0 {
		return s.cfg.UsageCleanup.BatchSize
	}
	return 5000
}

func (s *RequestLogCleanupService) workerInterval() time.Duration {
	if s == nil || s.cfg == nil {
		return 10 * time.Second
	}
	if s.cfg.UsageCleanup.WorkerIntervalSeconds > 0 {
		return time.Duration(s.cfg.UsageCleanup.WorkerIntervalSeconds) * time.Second
	}
	return 10 * time.Second
}

func (s *RequestLogCleanupService) taskTimeout() time.Duration {
	if s == nil || s.cfg == nil {
		return 30 * time.Minute
	}
	if s.cfg.UsageCleanup.TaskTimeoutSeconds > 0 {
		return time.Duration(s.cfg.UsageCleanup.TaskTimeoutSeconds) * time.Second
	}
	return 30 * time.Minute
}
