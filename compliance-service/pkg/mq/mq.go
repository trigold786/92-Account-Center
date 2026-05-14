package mq

import (
	"context"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

type MessageQueue interface {
	SendAuditLog(ctx context.Context, entry *model.AuditLogEntry) error
	StartConsumer(ctx context.Context, auditService service.AuditService) error
	Close() error
}
