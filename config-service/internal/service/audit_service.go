package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, operationType, operationObject, operator, result, details string) error
	ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error)
	GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error)
}

type auditService struct {
	auditRepo repository.AuditRepository
}

func NewAuditService(auditRepo repository.AuditRepository) AuditService {
	return &auditService{auditRepo: auditRepo}
}

func (s *auditService) Log(ctx context.Context, operationType, operationObject, operator, result, details string) error {
	entry := &model.AuditLog{
		OperationType:   operationType,
		OperationObject: operationObject,
		Operator:        operator,
		OperatorIP:      getClientIP(ctx),
		OperationResult: result,
		OperationDetails: details,
		SM3Hash:         "",
	}
	entry.SM3Hash = s.computeHash(entry)
	return s.auditRepo.CreateLog(ctx, entry)
}

func (s *auditService) computeHash(entry *model.AuditLog) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d",
		entry.OperationType, entry.OperationObject, entry.Operator,
		entry.OperatorIP, entry.OperationResult, entry.OperationDetails, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (s *auditService) ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error) {
	return s.auditRepo.ListLogs(ctx, filter)
}

func (s *auditService) GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error) {
	return s.auditRepo.GetLogByID(ctx, id)
}

// clientIPKey is used to store client IP in context
type clientIPKey struct{}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

func getClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value(clientIPKey{}).(string); ok {
		return ip
	}
	return ""
}
