package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

const (
	TypeRenewalReminder = "renewal:remind"
)

type RenewalWorker struct {
	svc *service.RenewalService
}

func NewRenewalWorker(svc *service.RenewalService) *RenewalWorker {
	return &RenewalWorker{svc: svc}
}

func (w *RenewalWorker) HandleReminder(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		UserID int64  `json:"user_id"`
		Channel string `json:"channel"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Sending renewal reminder: user=%d type=%s channel=%s", payload.UserID, payload.Type, payload.Channel)
	return nil
}

func NewRenewalReminderTask(userID int64, channel, reminderType string) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"channel": channel,
		"type":    reminderType,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeRenewalReminder, payload), nil
}
