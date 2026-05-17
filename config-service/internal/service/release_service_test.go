package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type MockReleaseRepository struct {
	mock.Mock
}

func (m *MockReleaseRepository) ListReleases(ctx context.Context, status string, page, pageSize int) ([]model.ConfigRelease, int, error) {
	args := m.Called(ctx, status, page, pageSize)
	return args.Get(0).([]model.ConfigRelease), args.Int(1), args.Error(2)
}

func (m *MockReleaseRepository) GetReleaseByID(ctx context.Context, id int64) (*model.ConfigRelease, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.ConfigRelease), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockReleaseRepository) CreateRelease(ctx context.Context, r *model.ConfigRelease) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockReleaseRepository) UpdateReleaseStatus(ctx context.Context, id int64, status, approvedBy string) error {
	args := m.Called(ctx, id, status, approvedBy)
	return args.Error(0)
}

func (m *MockReleaseRepository) ListReleaseItems(ctx context.Context, releaseID int64) ([]model.ConfigReleaseItem, error) {
	args := m.Called(ctx, releaseID)
	return args.Get(0).([]model.ConfigReleaseItem), args.Error(1)
}

func (m *MockReleaseRepository) AddReleaseItem(ctx context.Context, ri *model.ConfigReleaseItem) error {
	args := m.Called(ctx, ri)
	return args.Error(0)
}

func (m *MockReleaseRepository) RemoveReleaseItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestReleaseService_CreateRelease(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{Title: "Test Release", Description: "desc"}
	mockReleaseRepo.On("CreateRelease", mock.Anything, mock.MatchedBy(func(r *model.ConfigRelease) bool {
		return r.Status == "draft" && r.CreatedBy == "admin"
	})).Return(nil)
	mockAudit.On("Log", mock.Anything, "CREATE_RELEASE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.CreateRelease(context.Background(), rel, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "draft", rel.Status)
	mockReleaseRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_SubmitRelease_Success(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Title: "Test", Status: "draft"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)
	mockReleaseRepo.On("UpdateReleaseStatus", mock.Anything, int64(1), "pending", "").Return(nil)
	mockAudit.On("Log", mock.Anything, "SUBMIT_RELEASE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.SubmitRelease(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockReleaseRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_SubmitRelease_NotDraft(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "released"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)

	err := svc.SubmitRelease(context.Background(), 1, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in draft status")
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_SubmitRelease_NotFound(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.SubmitRelease(context.Background(), 999, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_ApproveRelease_Success(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "pending"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)
	mockReleaseRepo.On("UpdateReleaseStatus", mock.Anything, int64(1), "approved", "admin").Return(nil)
	mockAudit.On("Log", mock.Anything, "APPROVE_RELEASE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.ApproveRelease(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockReleaseRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_ApproveRelease_NotPending(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "draft"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)

	err := svc.ApproveRelease(context.Background(), 1, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending status")
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_RejectRelease_Success(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "pending"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)
	mockReleaseRepo.On("UpdateReleaseStatus", mock.Anything, int64(1), "rejected", "").Return(nil)
	mockAudit.On("Log", mock.Anything, "REJECT_RELEASE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.RejectRelease(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockReleaseRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_RejectRelease_NotPending(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "approved"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)

	err := svc.RejectRelease(context.Background(), 1, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in pending status")
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_ExecuteRelease_Success(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Title: "Release v1", Status: "approved"}
	items := []model.ConfigReleaseItem{
		{ID: 1, ItemID: 10, ValueBefore: "old", ValueAfter: "new"},
	}
	configItem := &model.ConfigItem{ID: 10, Code: "TEST_KEY", CurrentValue: "old"}

	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)
	mockReleaseRepo.On("ListReleaseItems", mock.Anything, int64(1)).Return(items, nil)
	mockConfigRepo.On("GetItemByID", mock.Anything, int64(10)).Return(configItem, nil)
	mockConfigRepo.On("UpdateItem", mock.Anything, mock.MatchedBy(func(item *model.ConfigItem) bool {
		return item.CurrentValue == "new"
	})).Return(nil)
	mockConfigRepo.On("CreateVersion", mock.Anything, mock.MatchedBy(func(v *model.ConfigVersion) bool {
		return v.ItemID == 10 && v.ValueBefore == "old" && v.ValueAfter == "new"
	})).Return(nil)
	mockReleaseRepo.On("UpdateReleaseStatus", mock.Anything, int64(1), "released", "").Return(nil)
	mockAudit.On("Log", mock.Anything, "EXECUTE_RELEASE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.ExecuteRelease(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockReleaseRepo.AssertExpectations(t)
	mockConfigRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_ExecuteRelease_NotApproved(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Status: "draft"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)

	err := svc.ExecuteRelease(context.Background(), 1, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in approved status")
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_AddReleaseItem_Success(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	configItem := &model.ConfigItem{ID: 10, Code: "TEST_KEY", CurrentValue: "current_val"}
	ri := &model.ConfigReleaseItem{ItemID: 10, ValueAfter: "new_val"}

	mockConfigRepo.On("GetItemByID", mock.Anything, int64(10)).Return(configItem, nil)
	mockReleaseRepo.On("AddReleaseItem", mock.Anything, mock.MatchedBy(func(r *model.ConfigReleaseItem) bool {
		return r.ValueBefore == "current_val"
	})).Return(nil)
	mockAudit.On("Log", mock.Anything, "ADD_RELEASE_ITEM", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.AddReleaseItem(context.Background(), ri, "admin")
	assert.NoError(t, err)
	mockReleaseRepo.AssertExpectations(t)
	mockConfigRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestReleaseService_AddReleaseItem_ItemNotFound(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	ri := &model.ConfigReleaseItem{ItemID: 999}
	mockConfigRepo.On("GetItemByID", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.AddReleaseItem(context.Background(), ri, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockReleaseRepo.AssertExpectations(t)
	mockConfigRepo.AssertExpectations(t)
}

func TestReleaseService_ListReleases(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	releases := []model.ConfigRelease{{ID: 1, Title: "Test", Status: "draft"}}
	mockReleaseRepo.On("ListReleases", mock.Anything, "draft", 1, 20).Return(releases, 1, nil)

	results, total, err := svc.ListReleases(context.Background(), "draft", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	mockReleaseRepo.AssertExpectations(t)
}

func TestReleaseService_GetReleaseByID(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewReleaseService(mockReleaseRepo, mockConfigRepo, mockAudit)

	rel := &model.ConfigRelease{ID: 1, Title: "Test"}
	mockReleaseRepo.On("GetReleaseByID", mock.Anything, int64(1)).Return(rel, nil)

	result, err := svc.GetReleaseByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	mockReleaseRepo.AssertExpectations(t)
}
