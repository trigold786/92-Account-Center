package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
)

type SubscriptionWorker struct {
	rdb       *redis.Client
	rebateSvc service.RebateService
	stream    string
	group     string
	consumer  string
}

func NewSubscriptionWorker(rdb *redis.Client, rebateSvc service.RebateService) *SubscriptionWorker {
	return &SubscriptionWorker{
		rdb:       rdb,
		rebateSvc: rebateSvc,
		stream:    "subscription:paid",
		group:     "credit-rebate-group",
		consumer:  "credit-worker-1",
	}
}

func (w *SubscriptionWorker) Start(ctx context.Context) {
	err := w.rdb.XGroupCreateMkStream(ctx, w.stream, w.group, "0").Err()
	if err != nil {
		log.Printf("XGroupCreateMkStream (may already exist): %v", err)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *SubscriptionWorker) processBatch(ctx context.Context) {
	streams, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.group,
		Consumer: w.consumer,
		Streams:  []string{w.stream, ">"},
		Count:    10,
		Block:    0,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		log.Printf("XReadGroup error: %v", err)
		return
	}

	for _, s := range streams {
		for _, msg := range s.Messages {
			w.handleMessage(ctx, msg.ID, msg.Values)
		}
	}
}

func (w *SubscriptionWorker) handleMessage(ctx context.Context, msgID string, values map[string]interface{}) {
	payload, ok := values["payload"].(string)
	if !ok {
		log.Printf("invalid message format: %s", msgID)
		w.rdb.XAck(ctx, w.stream, w.group, msgID)
		return
	}

	var event model.ProcessSubscriptionPaidEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("failed to unmarshal event: %v", err)
		w.rdb.XAck(ctx, w.stream, w.group, msgID)
		return
	}

	if err := w.rebateSvc.ProcessSubscriptionPaid(ctx, &event); err != nil {
		log.Printf("failed to process rebate for msg %s: %v", msgID, err)
		return
	}

	w.rdb.XAck(ctx, w.stream, w.group, msgID)
}
