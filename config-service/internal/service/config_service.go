package service

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
)

type ConfigService interface {
	ListGroups(ctx context.Context) ([]model.ConfigGroup, error)
	GetGroupByID(ctx context.Context, id int64) (*model.ConfigGroup, error)
	CreateGroup(ctx context.Context, g *model.ConfigGroup, operator string) error
	UpdateGroup(ctx context.Context, g *model.ConfigGroup, operator string) error
	DeleteGroup(ctx context.Context, id int64, operator string) error

	ListItems(ctx context.Context, filter model.ConfigItemFilter) ([]model.ConfigItem, int, error)
	GetItemByID(ctx context.Context, id int64) (*model.ConfigItem, error)
	GetItemByCode(ctx context.Context, code string) (*model.ConfigItem, error)
	CreateItem(ctx context.Context, item *model.ConfigItem, operator string) error
	UpdateItem(ctx context.Context, item *model.ConfigItem, changeReason, operator string) error
	DeleteItem(ctx context.Context, id int64, operator string) error
	ResetItemToDefault(ctx context.Context, id int64, operator string) error

	ListVersionsByItemID(ctx context.Context, itemID int64) ([]model.ConfigVersion, error)
	GetTotalCount(ctx context.Context) (int, error)
}

type configService struct {
	configRepo  repository.ConfigRepository
	auditSvc    AuditService
}

func NewConfigService(configRepo repository.ConfigRepository, auditSvc AuditService) ConfigService {
	return &configService{
		configRepo: configRepo,
		auditSvc:   auditSvc,
	}
}

func (s *configService) ListGroups(ctx context.Context) ([]model.ConfigGroup, error) {
	return s.configRepo.ListGroups(ctx)
}

func (s *configService) GetGroupByID(ctx context.Context, id int64) (*model.ConfigGroup, error) {
	return s.configRepo.GetGroupByID(ctx, id)
}

func (s *configService) CreateGroup(ctx context.Context, g *model.ConfigGroup, operator string) error {
	if err := s.configRepo.CreateGroup(ctx, g); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "CREATE_GROUP", fmt.Sprintf("config_groups:%d", g.ID), operator, "success",
		fmt.Sprintf("Created group: %s", g.Name))
	return nil
}

func (s *configService) UpdateGroup(ctx context.Context, g *model.ConfigGroup, operator string) error {
	if err := s.configRepo.UpdateGroup(ctx, g); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "UPDATE_GROUP", fmt.Sprintf("config_groups:%d", g.ID), operator, "success",
		fmt.Sprintf("Updated group: %s", g.Name))
	return nil
}

func (s *configService) DeleteGroup(ctx context.Context, id int64, operator string) error {
	if err := s.configRepo.DeleteGroup(ctx, id); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "DELETE_GROUP", fmt.Sprintf("config_groups:%d", id), operator, "success", "")
	return nil
}

func (s *configService) ListItems(ctx context.Context, filter model.ConfigItemFilter) ([]model.ConfigItem, int, error) {
	return s.configRepo.ListItems(ctx, filter)
}

func (s *configService) GetItemByID(ctx context.Context, id int64) (*model.ConfigItem, error) {
	return s.configRepo.GetItemByID(ctx, id)
}

func (s *configService) GetItemByCode(ctx context.Context, code string) (*model.ConfigItem, error) {
	return s.configRepo.GetItemByCode(ctx, code)
}

func (s *configService) CreateItem(ctx context.Context, item *model.ConfigItem, operator string) error {
	if err := s.configRepo.CreateItem(ctx, item); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "CREATE_ITEM", fmt.Sprintf("config_items:%d", item.ID), operator, "success",
		fmt.Sprintf("Created item: %s (%s)", item.Code, item.Name))
	return nil
}

func (s *configService) UpdateItem(ctx context.Context, item *model.ConfigItem, changeReason, operator string) error {
	old, err := s.configRepo.GetItemByID(ctx, item.ID)
	if err != nil {
		return err
	}
	if old == nil {
		return fmt.Errorf("config item not found: %d", item.ID)
	}

	oldValue := old.CurrentValue
	if err := s.configRepo.UpdateItem(ctx, item); err != nil {
		return err
	}

	s.configRepo.CreateVersion(ctx, &model.ConfigVersion{
		ItemID:       item.ID,
		ValueBefore:  oldValue,
		ValueAfter:   item.CurrentValue,
		ChangeReason: changeReason,
		ChangedBy:    operator,
	})

	s.auditSvc.Log(ctx, "UPDATE_ITEM", fmt.Sprintf("config_items:%d", item.ID), operator, "success",
		fmt.Sprintf("Updated %s: %s -> %s", item.Code, oldValue, item.CurrentValue))
	return nil
}

func (s *configService) DeleteItem(ctx context.Context, id int64, operator string) error {
	if err := s.configRepo.DeleteItem(ctx, id); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "DELETE_ITEM", fmt.Sprintf("config_items:%d", id), operator, "success", "")
	return nil
}

func (s *configService) ResetItemToDefault(ctx context.Context, id int64, operator string) error {
	item, err := s.configRepo.GetItemByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf("config item not found: %d", id)
	}

	oldValue := item.CurrentValue
	if err := s.configRepo.ResetItemToDefault(ctx, id); err != nil {
		return err
	}

	s.configRepo.CreateVersion(ctx, &model.ConfigVersion{
		ItemID:       id,
		ValueBefore:  oldValue,
		ValueAfter:   item.DefaultValue,
		ChangeReason: "reset to default",
		ChangedBy:    operator,
	})

	s.auditSvc.Log(ctx, "RESET_ITEM", fmt.Sprintf("config_items:%d", id), operator, "success",
		fmt.Sprintf("Reset %s to default: %s -> %s", item.Code, oldValue, item.DefaultValue))
	return nil
}

func (s *configService) ListVersionsByItemID(ctx context.Context, itemID int64) ([]model.ConfigVersion, error) {
	return s.configRepo.ListVersionsByItemID(ctx, itemID)
}

func (s *configService) GetTotalCount(ctx context.Context) (int, error) {
	return s.configRepo.GetTotalCount(ctx)
}
