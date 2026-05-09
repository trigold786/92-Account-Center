package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"

	"account-center/audit-log-service/internal/model"
	"account-center/audit-log-service/internal/service"
)

type AuditLogConsumer struct {
	consumerGroup sarama.ConsumerGroup
	auditService  service.AuditService
	topic         string
	ready         chan bool
}

func NewAuditLogConsumer(brokers []string, groupID, topic string, auditService service.AuditService) (*AuditLogConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &AuditLogConsumer{
		consumerGroup: consumerGroup,
		auditService:  auditService,
		topic:         topic,
		ready:         make(chan bool),
	}, nil
}

func (c *AuditLogConsumer) Start(ctx context.Context) error {
	go func() {
		for {
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, c); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready
	log.Printf("Audit log consumer started, listening to topic: %s", c.topic)
	return nil
}

func (c *AuditLogConsumer) Close() error {
	return c.consumerGroup.Close()
}

func (c *AuditLogConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *AuditLogConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *AuditLogConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("Message channel was closed")
				return nil
			}

			var entry model.AuditLogEntry
			if err := json.Unmarshal(message.Value, &entry); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			_, err := c.auditService.RecordLog(session.Context(), &entry)
			if err != nil {
				log.Printf("Failed to record audit log: %v", err)
			} else {
				log.Printf("Recorded audit log from Kafka: action=%s, resource=%s",
					entry.ActionType, entry.TargetResource)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
