package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/model"
)

type AuditLogProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewAuditLogProducer(brokers []string, topic string) (*AuditLogProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 10 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &AuditLogProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *AuditLogProducer) SendAuditLog(ctx context.Context, entry *model.AuditLogEntry) error {
	value, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log entry: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(value),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	return nil
}

func (p *AuditLogProducer) Close() error {
	return p.producer.Close()
}
