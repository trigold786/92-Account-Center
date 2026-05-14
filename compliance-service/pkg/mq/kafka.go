package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

type KafkaMQ struct {
	brokers       []string
	topic         string
	groupID       string
	producer      sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup
	ready         chan bool
}

func NewKafkaMQ(brokers []string, topic, groupID string) (*KafkaMQ, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 10 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("kafka: create producer: %w", err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, consumerConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("kafka: create consumer: %w", err)
	}

	return &KafkaMQ{
		brokers:       brokers,
		topic:         topic,
		groupID:       groupID,
		producer:      producer,
		consumerGroup: consumerGroup,
		ready:         make(chan bool),
	}, nil
}

func (k *KafkaMQ) SendAuditLog(ctx context.Context, entry *model.AuditLogEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	value, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("kafka: marshal: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.ByteEncoder(value),
	}

	_, _, err = k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("kafka: send: %w", err)
	}
	return nil
}

func (k *KafkaMQ) StartConsumer(ctx context.Context, auditService service.AuditService) error {
	handler := &kafkaConsumerHandler{
		auditService: auditService,
		ready:        k.ready,
	}

	go func() {
		for {
			if err := k.consumerGroup.Consume(ctx, []string{k.topic}, handler); err != nil {
				log.Printf("Kafka consume error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			k.ready = make(chan bool)
		}
	}()

	<-k.ready
	log.Printf("Kafka consumer started: topic=%s group=%s", k.topic, k.groupID)
	return nil
}

func (k *KafkaMQ) Close() error {
	prodErr := k.producer.Close()
	consErr := k.consumerGroup.Close()
	if prodErr != nil {
		return prodErr
	}
	return consErr
}

type kafkaConsumerHandler struct {
	auditService service.AuditService
	ready        chan bool
}

func (h *kafkaConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *kafkaConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *kafkaConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			var entry model.AuditLogEntry
			if err := json.Unmarshal(message.Value, &entry); err != nil {
				log.Printf("Kafka: unmarshal error: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			if _, err := h.auditService.RecordLog(session.Context(), &entry); err != nil {
				log.Printf("Kafka: record error: %v", err)
			} else {
				log.Printf("Recorded audit log from Kafka: action=%s resource=%s",
					entry.ActionType, entry.TargetResource)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
