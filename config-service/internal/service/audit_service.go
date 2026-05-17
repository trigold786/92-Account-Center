package service

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/config-service/internal/crypto"
	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, operationType, operationObject, operator, result, details string) error
	ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error)
	GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error)
}

type auditService struct {
	auditRepo  repository.AuditRepository
	lastHash   string
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
	}
	entry.SM3Hash = s.computeHash(entry)
	if err := s.auditRepo.CreateLog(ctx, entry); err != nil {
		return err
	}
	s.lastHash = entry.SM3Hash
	return nil
}

func (s *auditService) computeHash(entry *model.AuditLog) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		entry.OperationType, entry.OperationObject, entry.Operator,
		entry.OperatorIP, entry.OperationResult, entry.OperationDetails, s.lastHash)
	return crypto.SM3Hash([]byte(data))
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
