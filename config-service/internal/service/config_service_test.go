package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) ListGroups(ctx context.Context) ([]model.ConfigGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.ConfigGroup), args.Error(1)
}

func (m *MockConfigRepository) GetGroupByID(ctx context.Context, id int64) (*model.ConfigGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.ConfigGroup), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockConfigRepository) CreateGroup(ctx context.Context, g *model.ConfigGroup) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockConfigRepository) UpdateGroup(ctx context.Context, g *model.ConfigGroup) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockConfigRepository) DeleteGroup(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConfigRepository) ListItems(ctx context.Context, filter model.ConfigItemFilter) ([]model.ConfigItem, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.ConfigItem), args.Int(1), args.Error(2)
}

func (m *MockConfigRepository) GetItemByID(ctx context.Context, id int64) (*model.ConfigItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.ConfigItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockConfigRepository) GetItemByCode(ctx context.Context, code string) (*model.ConfigItem, error) {
	args := m.Called(ctx, code)
	if args.Get(0) != nil {
		return args.Get(0).(*model.ConfigItem), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockConfigRepository) CreateItem(ctx context.Context, item *model.ConfigItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockConfigRepository) UpdateItem(ctx context.Context, item *model.ConfigItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockConfigRepository) DeleteItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConfigRepository) ResetItemToDefault(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConfigRepository) GetTotalCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockConfigRepository) ListVersionsByItemID(ctx context.Context, itemID int64) ([]model.ConfigVersion, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]model.ConfigVersion), args.Error(1)
}

func (m *MockConfigRepository) CreateVersion(ctx context.Context, v *model.ConfigVersion) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}

type MockAuditSvc struct {
	mock.Mock
}

func (m *MockAuditSvc) Log(ctx context.Context, opType, opObj, operator, result, details string) error {
	args := m.Called(ctx, opType, opObj, operator, result, details)
	return args.Error(0)
}

func (m *MockAuditSvc) ListLogs(ctx context.Context, filter model.AuditLogFilter) ([]model.AuditLog, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.AuditLog), args.Int(1), args.Error(2)
}

func (m *MockAuditSvc) GetLogByID(ctx context.Context, id int64) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.AuditLog), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestConfigService_ListGroups(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	expected := []model.ConfigGroup{{ID: 1, Name: "test"}}
	mockRepo.On("ListGroups", mock.Anything).Return(expected, nil)

	groups, err := svc.ListGroups(context.Background())
	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	mockRepo.AssertExpectations(t)
}

func TestConfigService_GetGroupByID(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	expected := &model.ConfigGroup{ID: 1, Name: "test"}
	mockRepo.On("GetGroupByID", mock.Anything, int64(1)).Return(expected, nil)

	group, err := svc.GetGroupByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "test", group.Name)
	mockRepo.AssertExpectations(t)
}

func TestConfigService_CreateGroup(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	group := &model.ConfigGroup{Name: "new-group"}
	mockRepo.On("CreateGroup", mock.Anything, group).Return(nil)
	mockAudit.On("Log", mock.Anything, "CREATE_GROUP", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.CreateGroup(context.Background(), group, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_UpdateGroup(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	group := &model.ConfigGroup{ID: 1, Name: "updated"}
	mockRepo.On("UpdateGroup", mock.Anything, group).Return(nil)
	mockAudit.On("Log", mock.Anything, "UPDATE_GROUP", "config_groups:1", "admin", "success", mock.Anything).Return(nil)

	err := svc.UpdateGroup(context.Background(), group, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_DeleteGroup(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	mockRepo.On("DeleteGroup", mock.Anything, int64(1)).Return(nil)
	mockAudit.On("Log", mock.Anything, "DELETE_GROUP", "config_groups:1", "admin", "success", mock.Anything).Return(nil)

	err := svc.DeleteGroup(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_DeleteGroup_Error(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	mockRepo.On("DeleteGroup", mock.Anything, int64(1)).Return(assert.AnError)

	err := svc.DeleteGroup(context.Background(), 1, "admin")
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestConfigService_CreateItem(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	item := &model.ConfigItem{Code: "TEST_KEY", Name: "Test", GroupID: 1, DataType: "STRING"}
	mockRepo.On("CreateItem", mock.Anything, item).Return(nil)
	mockAudit.On("Log", mock.Anything, "CREATE_ITEM", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.CreateItem(context.Background(), item, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_UpdateItem_Success(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	oldItem := &model.ConfigItem{ID: 1, Code: "TEST_KEY", CurrentValue: "old_val"}
	updatedItem := &model.ConfigItem{ID: 1, Code: "TEST_KEY", CurrentValue: "new_val"}

	mockRepo.On("GetItemByID", mock.Anything, int64(1)).Return(oldItem, nil)
	mockRepo.On("UpdateItem", mock.Anything, updatedItem).Return(nil)
	mockRepo.On("CreateVersion", mock.Anything, mock.MatchedBy(func(v *model.ConfigVersion) bool {
		return v.ItemID == 1 && v.ValueBefore == "old_val" && v.ValueAfter == "new_val"
	})).Return(nil)
	mockAudit.On("Log", mock.Anything, "UPDATE_ITEM", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.UpdateItem(context.Background(), updatedItem, "test change", "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_UpdateItem_NotFound(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	item := &model.ConfigItem{ID: 999, CurrentValue: "new"}

	mockRepo.On("GetItemByID", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.UpdateItem(context.Background(), item, "reason", "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestConfigService_DeleteItem(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	mockRepo.On("DeleteItem", mock.Anything, int64(1)).Return(nil)
	mockAudit.On("Log", mock.Anything, "DELETE_ITEM", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.DeleteItem(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_ResetItemToDefault_Success(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	item := &model.ConfigItem{ID: 1, Code: "TEST_KEY", CurrentValue: "changed", DefaultValue: "default"}
	mockRepo.On("GetItemByID", mock.Anything, int64(1)).Return(item, nil)
	mockRepo.On("ResetItemToDefault", mock.Anything, int64(1)).Return(nil)
	mockRepo.On("CreateVersion", mock.Anything, mock.MatchedBy(func(v *model.ConfigVersion) bool {
		return v.ValueBefore == "changed" && v.ValueAfter == "default"
	})).Return(nil)
	mockAudit.On("Log", mock.Anything, "RESET_ITEM", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.ResetItemToDefault(context.Background(), 1, "admin")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestConfigService_ResetItemToDefault_NotFound(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	mockRepo.On("GetItemByID", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.ResetItemToDefault(context.Background(), 999, "admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestConfigService_GetItemByCode(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	expected := &model.ConfigItem{ID: 1, Code: "TEST_KEY"}
	mockRepo.On("GetItemByCode", mock.Anything, "TEST_KEY").Return(expected, nil)

	item, err := svc.GetItemByCode(context.Background(), "TEST_KEY")
	assert.NoError(t, err)
	assert.Equal(t, "TEST_KEY", item.Code)
	mockRepo.AssertExpectations(t)
}

func TestConfigService_GetTotalCount(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	mockRepo.On("GetTotalCount", mock.Anything).Return(106, nil)

	count, err := svc.GetTotalCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 106, count)
	mockRepo.AssertExpectations(t)
}

func TestConfigService_ListItems(t *testing.T) {
	mockRepo := new(MockConfigRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewConfigService(mockRepo, mockAudit)

	filter := model.ConfigItemFilter{Page: 1, PageSize: 20}
	items := []model.ConfigItem{{ID: 1, Code: "TEST"}}
	mockRepo.On("ListItems", mock.Anything, filter).Return(items, 1, nil)

	results, total, err := svc.ListItems(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	mockRepo.AssertExpectations(t)
}
