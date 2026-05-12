package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/model"
	"github.com/trigold786/92-Account-Center/audit-log-service/internal/service"
)

type RedisStreamsMQ struct {
	client     *redis.Client
	streamKey  string
	groupName  string
	consumerID string
}

func NewRedisStreamsMQ(redisAddr, redisPassword string, redisDB int, streamKey, groupName, consumerID string) *RedisStreamsMQ {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	return &RedisStreamsMQ{
		client:     rdb,
		streamKey:  streamKey,
		groupName:  groupName,
		consumerID: consumerID,
	}
}

func (r *RedisStreamsMQ) SendAuditLog(ctx context.Context, entry *model.AuditLogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("redis streams: marshal: %w", err)
	}

	values := map[string]interface{}{
		"payload": string(data),
	}

	if err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.streamKey,
		Values: values,
	}).Err(); err != nil {
		return fmt.Errorf("redis streams: xadd: %w", err)
	}

	return nil
}

func (r *RedisStreamsMQ) StartConsumer(ctx context.Context, auditService service.AuditService) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis streams: ping: %w", err)
	}

	if err := r.client.XGroupCreateMkStream(ctx, r.streamKey, r.groupName, "$").Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("redis streams: create group: %w", err)
		}
	}

	go r.consumeLoop(ctx, auditService)

	log.Printf("Redis Streams consumer started: stream=%s group=%s consumer=%s",
		r.streamKey, r.groupName, r.consumerID)
	return nil
}

func (r *RedisStreamsMQ) consumeLoop(ctx context.Context, auditService service.AuditService) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		results, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    r.groupName,
			Consumer: r.consumerID,
			Streams:  []string{r.streamKey, ">"},
			Count:    10,
			Block:    2 * time.Second,
		}).Result()

		if err != nil {
			if err != redis.Nil {
				log.Printf("Redis Streams read error: %v", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, stream := range results {
			for _, msg := range stream.Messages {
				r.processMessage(ctx, auditService, msg.ID, msg.Values)
			}
		}
	}
}

func (r *RedisStreamsMQ) processMessage(ctx context.Context, auditService service.AuditService, msgID string, values map[string]interface{}) {
	payload, ok := values["payload"].(string)
	if !ok {
		log.Printf("Redis Streams: invalid message payload format, acking %s", msgID)
		r.ackMessage(ctx, msgID)
		return
	}

	var entry model.AuditLogEntry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		log.Printf("Redis Streams: unmarshal error: %v, acking %s", err, msgID)
		r.ackMessage(ctx, msgID)
		return
	}

	if _, err := auditService.RecordLog(ctx, &entry); err != nil {
		log.Printf("Redis Streams: record error: %v, message %s will be redelivered", err, msgID)
		return
	}

	log.Printf("Recorded audit log from Redis Streams: action=%s resource=%s",
		entry.ActionType, entry.TargetResource)
	r.ackMessage(ctx, msgID)
}

func (r *RedisStreamsMQ) ackMessage(ctx context.Context, msgID string) {
	if err := r.client.XAck(ctx, r.streamKey, r.groupName, msgID).Err(); err != nil {
		log.Printf("Redis Streams: ack error for message %s: %v", msgID, err)
	}
}

func (r *RedisStreamsMQ) Close() error {
	return r.client.Close()
}
