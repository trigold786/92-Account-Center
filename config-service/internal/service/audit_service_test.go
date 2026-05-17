package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.AuditLog), args.Int(1), args.Error(2)
}

func (m *MockAuditRepository) GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.AuditLog), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuditRepository) CreateLog(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func TestAuditService_Log_Success(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	mockRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.OperationType == "TEST_OP" && l.Operator == "admin" && l.SM3Hash != ""
	})).Return(nil).Once()

	ctx := WithClientIP(context.Background(), "10.0.0.1")
	err := svc.Log(ctx, "TEST_OP", "test:1", "admin", "success", "test details")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_Log_WithoutIP(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	mockRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.OperatorIP == ""
	})).Return(nil).Once()

	err := svc.Log(context.Background(), "TEST_OP", "test:1", "admin", "success", "")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_Log_HashChain(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	var hashes []string
	mockRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		hashes = append(hashes, l.SM3Hash)
		return true
	})).Return(nil).Times(2)

	err := svc.Log(context.Background(), "OP1", "obj:1", "user1", "success", "first")
	assert.NoError(t, err)

	err = svc.Log(context.Background(), "OP2", "obj:2", "user2", "success", "second")
	assert.NoError(t, err)

	assert.Len(t, hashes, 2)
	assert.NotEqual(t, hashes[0], hashes[1], "hash chain should produce different hashes")
	mockRepo.AssertExpectations(t)
}

func TestAuditService_ListLogs(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	filter := model.AuditLogFilter{Page: 1, PageSize: 10}
	expectedLogs := []model.AuditLog{{ID: 1, OperationType: "TEST"}}
	mockRepo.On("ListLogs", mock.Anything, filter).Return(expectedLogs, 1, nil)

	logs, total, err := svc.ListLogs(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, logs, 1)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetLogByID(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	expected := &model.AuditLog{ID: 1, OperationType: "TEST"}
	mockRepo.On("GetLogByID", mock.Anything, int64(1)).Return(expected, nil)

	log, err := svc.GetLogByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), log.ID)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_GetLogByID_NotFound(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	svc := NewAuditService(mockRepo)

	mockRepo.On("GetLogByID", mock.Anything, int64(999)).Return(nil, nil)

	log, err := svc.GetLogByID(context.Background(), 999)
	assert.NoError(t, err)
	assert.Nil(t, log)
	mockRepo.AssertExpectations(t)
}
