package service

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
)

type ReleaseService interface {
	ListReleases(ctx context.Context, status string, page, pageSize int) ([]model.ConfigRelease, int, error)
	GetReleaseByID(ctx context.Context, id int64) (*model.ConfigRelease, error)
	CreateRelease(ctx context.Context, rel *model.ConfigRelease, operator string) error
	SubmitRelease(ctx context.Context, id int64, operator string) error
	ApproveRelease(ctx context.Context, id int64, operator string) error
	RejectRelease(ctx context.Context, id int64, operator string) error
	ExecuteRelease(ctx context.Context, id int64, operator string) error

	ListReleaseItems(ctx context.Context, releaseID int64) ([]model.ConfigReleaseItem, error)
	AddReleaseItem(ctx context.Context, ri *model.ConfigReleaseItem, operator string) error
	RemoveReleaseItem(ctx context.Context, id int64, operator string) error
}

type releaseService struct {
	releaseRepo repository.ReleaseRepository
	configRepo  repository.ConfigRepository
	auditSvc    AuditService
}

func NewReleaseService(releaseRepo repository.ReleaseRepository, configRepo repository.ConfigRepository, auditSvc AuditService) ReleaseService {
	return &releaseService{
		releaseRepo: releaseRepo,
		configRepo:  configRepo,
		auditSvc:    auditSvc,
	}
}

func (s *releaseService) ListReleases(ctx context.Context, status string, page, pageSize int) ([]model.ConfigRelease, int, error) {
	return s.releaseRepo.ListReleases(ctx, status, page, pageSize)
}

func (s *releaseService) GetReleaseByID(ctx context.Context, id int64) (*model.ConfigRelease, error) {
	return s.releaseRepo.GetReleaseByID(ctx, id)
}

func (s *releaseService) CreateRelease(ctx context.Context, rel *model.ConfigRelease, operator string) error {
	rel.Status = "draft"
	rel.CreatedBy = operator
	if err := s.releaseRepo.CreateRelease(ctx, rel); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "CREATE_RELEASE", fmt.Sprintf("config_releases:%d", rel.ID), operator, "success",
		fmt.Sprintf("Created release: %s", rel.Title))
	return nil
}

func (s *releaseService) SubmitRelease(ctx context.Context, id int64, operator string) error {
	rel, err := s.releaseRepo.GetReleaseByID(ctx, id)
	if err != nil {
		return err
	}
	if rel == nil {
		return fmt.Errorf("release not found: %d", id)
	}
	if rel.Status != "draft" {
		return fmt.Errorf("release %d is not in draft status (current: %s)", id, rel.Status)
	}
	if err := s.releaseRepo.UpdateReleaseStatus(ctx, id, "pending", ""); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "SUBMIT_RELEASE", fmt.Sprintf("config_releases:%d", id), operator, "success", "")
	return nil
}

func (s *releaseService) ApproveRelease(ctx context.Context, id int64, operator string) error {
	rel, err := s.releaseRepo.GetReleaseByID(ctx, id)
	if err != nil {
		return err
	}
	if rel == nil {
		return fmt.Errorf("release not found: %d", id)
	}
	if rel.Status != "pending" {
		return fmt.Errorf("release %d is not in pending status (current: %s)", id, rel.Status)
	}
	if err := s.releaseRepo.UpdateReleaseStatus(ctx, id, "approved", operator); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "APPROVE_RELEASE", fmt.Sprintf("config_releases:%d", id), operator, "success", "")
	return nil
}

func (s *releaseService) RejectRelease(ctx context.Context, id int64, operator string) error {
	rel, err := s.releaseRepo.GetReleaseByID(ctx, id)
	if err != nil {
		return err
	}
	if rel == nil {
		return fmt.Errorf("release not found: %d", id)
	}
	if rel.Status != "pending" {
		return fmt.Errorf("release %d is not in pending status (current: %s)", id, rel.Status)
	}
	if err := s.releaseRepo.UpdateReleaseStatus(ctx, id, "rejected", ""); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "REJECT_RELEASE", fmt.Sprintf("config_releases:%d", id), operator, "success", "")
	return nil
}

func (s *releaseService) ExecuteRelease(ctx context.Context, id int64, operator string) error {
	rel, err := s.releaseRepo.GetReleaseByID(ctx, id)
	if err != nil {
		return err
	}
	if rel == nil {
		return fmt.Errorf("release not found: %d", id)
	}
	if rel.Status != "approved" {
		return fmt.Errorf("release %d is not in approved status (current: %s)", id, rel.Status)
	}

	items, err := s.releaseRepo.ListReleaseItems(ctx, id)
	if err != nil {
		return err
	}

	for _, ri := range items {
		item, err := s.configRepo.GetItemByID(ctx, ri.ItemID)
		if err != nil {
			return fmt.Errorf("failed to get item %d: %w", ri.ItemID, err)
		}
		if item == nil {
			continue
		}

		item.CurrentValue = ri.ValueAfter
		if err := s.configRepo.UpdateItem(ctx, item); err != nil {
			return fmt.Errorf("failed to update item %d: %w", ri.ItemID, err)
		}

		s.configRepo.CreateVersion(ctx, &model.ConfigVersion{
			ItemID:       ri.ItemID,
			ValueBefore:  ri.ValueBefore,
			ValueAfter:   ri.ValueAfter,
			ChangeReason: "release: " + rel.Title,
			ChangedBy:    operator,
		})
	}

	if err := s.releaseRepo.UpdateReleaseStatus(ctx, id, "released", ""); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "EXECUTE_RELEASE", fmt.Sprintf("config_releases:%d", id), operator, "success",
		fmt.Sprintf("Executed release with %d items", len(items)))
	return nil
}

func (s *releaseService) ListReleaseItems(ctx context.Context, releaseID int64) ([]model.ConfigReleaseItem, error) {
	return s.releaseRepo.ListReleaseItems(ctx, releaseID)
}

func (s *releaseService) AddReleaseItem(ctx context.Context, ri *model.ConfigReleaseItem, operator string) error {
	item, err := s.configRepo.GetItemByID(ctx, ri.ItemID)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf("config item not found: %d", ri.ItemID)
	}
	ri.ValueBefore = item.CurrentValue

	if err := s.releaseRepo.AddReleaseItem(ctx, ri); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "ADD_RELEASE_ITEM", fmt.Sprintf("config_release_items:%d", ri.ID), operator, "success", "")
	return nil
}

func (s *releaseService) RemoveReleaseItem(ctx context.Context, id int64, operator string) error {
	if err := s.releaseRepo.RemoveReleaseItem(ctx, id); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "REMOVE_RELEASE_ITEM", "", operator, "success", "")
	return nil
}
