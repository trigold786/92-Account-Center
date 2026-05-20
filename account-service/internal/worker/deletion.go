package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"

	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

const (
	TaskProcessDeletions = "deletion:process_expired"
	SchedulePeriod       = 1 * time.Hour
)

type DeletionWorker struct {
	svc    service.DeletionService
	logger *slog.Logger
}

func NewDeletionWorker(svc service.DeletionService, logger *slog.Logger) *DeletionWorker {
	return &DeletionWorker{svc: svc, logger: logger}
}

func (w *DeletionWorker) HandleTask(ctx context.Context, t *asynq.Task) error {
	w.logger.Info("processing expired account deletions")

	count, err := w.svc.ProcessExpiredDeletions(ctx)
	if err != nil {
		w.logger.Error("deletion processing failed", "error", err.Error())
		return fmt.Errorf("process expired deletions: %w", err)
	}

	w.logger.Info("deletion processing complete", "accounts_anonymized", count)
	return nil
}

func NewScheduler(redisAddr string) (*asynq.Scheduler, error) {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		nil,
	)

	task, err := json.Marshal(map[string]string{"trigger": "scheduled"})
	if err != nil {
		return nil, fmt.Errorf("marshal scheduled task: %w", err)
	}

	_, err = scheduler.Register("@every 1h", asynq.NewTask(TaskProcessDeletions, task))
	if err != nil {
		return nil, fmt.Errorf("register periodic task: %w", err)
	}

	return scheduler, nil
}

func NewServeMux(w *DeletionWorker) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskProcessDeletions, w.HandleTask)
	return mux
}
