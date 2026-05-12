package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/kyb-service/internal/model"
	"github.com/trigold786/92-Account-Center/kyb-service/internal/repository"
)

var (
	ErrSubAccountNotFound   = errors.New("sub-account not found")
	ErrNotEnterpriseOwner   = errors.New("only enterprise owner can manage sub-accounts")
	ErrLivenessRequired     = errors.New("liveness verification required before accessing sensitive data")
)

type SubAccountRepository interface {
	CreateSubAccount(ctx context.Context, sa *model.SubAccount) error
	GetSubAccount(ctx context.Context, id uuid.UUID) (*model.SubAccount, error)
	ListSubAccounts(ctx context.Context, enterpriseID uuid.UUID) ([]*model.SubAccount, error)
	UpdateSubAccount(ctx context.Context, sa *model.SubAccount) error
	DeleteSubAccount(ctx context.Context, id uuid.UUID) error
}

type SubAccountService interface {
	CreateSubAccount(ctx context.Context, ownerUserID string, req *model.CreateSubAccountRequest) (*model.SubAccountResponse, error)
	UpdateSubAccount(ctx context.Context, subAccountID string, req *model.UpdateSubAccountRequest) (*model.SubAccountResponse, error)
	ListSubAccounts(ctx context.Context, enterpriseID string) ([]*model.SubAccountResponse, error)
	DeleteSubAccount(ctx context.Context, subAccountID string) error
	RequireLiveness(ctx context.Context, req *model.RequireLivenessRequest) error
	CompleteLiveness(ctx context.Context, req *model.CompleteLivenessRequest) error
	CheckLivenessRequired(ctx context.Context, userID string, accessType string) (bool, error)
}

type subAccountService struct {
	subRepo   SubAccountRepository
	entRepo   repository.EnterpriseRepository
}

func NewSubAccountService(subRepo SubAccountRepository, entRepo repository.EnterpriseRepository) SubAccountService {
	return &subAccountService{subRepo: subRepo, entRepo: entRepo}
}

func (s *subAccountService) CreateSubAccount(ctx context.Context, ownerUserID string, req *model.CreateSubAccountRequest) (*model.SubAccountResponse, error) {
	ownerUID, err := uuid.Parse(ownerUserID)
	if err != nil {
		return nil, errors.New("invalid owner user ID")
	}

	entID, err := uuid.Parse(req.EnterpriseID)
	if err != nil {
		return nil, errors.New("invalid enterprise ID")
	}

	ent, err := s.entRepo.GetByID(ctx, entID)
	if err != nil || ent == nil {
		return nil, ErrEnterpriseNotFound
	}

	if ent.UserID != ownerUID {
		return nil, ErrNotEnterpriseOwner
	}

	if ent.VerificationStatus != model.VerificationStatusVerified {
		return nil, errors.New("enterprise must be verified before creating sub-accounts")
	}

	subUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid sub-account user ID")
	}

	role := model.SubAccountRole(req.Role)
	if role != model.SubAccountRoleAdmin && role != model.SubAccountRoleMember {
		role = model.SubAccountRoleMember
	}

	now := time.Now()
	sa := &model.SubAccount{
		ID:           uuid.New(),
		EnterpriseID: entID,
		UserID:       subUID,
		Role:         role,
		Status:       model.SubAccountStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.subRepo.CreateSubAccount(ctx, sa); err != nil {
		return nil, err
	}

	return toSubAccountResponse(sa), nil
}

func (s *subAccountService) UpdateSubAccount(ctx context.Context, subAccountID string, req *model.UpdateSubAccountRequest) (*model.SubAccountResponse, error) {
	saID, err := uuid.Parse(subAccountID)
	if err != nil {
		return nil, errors.New("invalid sub-account ID")
	}

	sa, err := s.subRepo.GetSubAccount(ctx, saID)
	if err != nil || sa == nil {
		return nil, ErrSubAccountNotFound
	}

	if req.Role != "" {
		sa.Role = model.SubAccountRole(req.Role)
	}
	if req.Status != "" {
		sa.Status = model.SubAccountStatus(req.Status)
	}
	sa.UpdatedAt = time.Now()

	if err := s.subRepo.UpdateSubAccount(ctx, sa); err != nil {
		return nil, err
	}

	return toSubAccountResponse(sa), nil
}

func (s *subAccountService) ListSubAccounts(ctx context.Context, enterpriseID string) ([]*model.SubAccountResponse, error) {
	entID, err := uuid.Parse(enterpriseID)
	if err != nil {
		return nil, errors.New("invalid enterprise ID")
	}

	subs, err := s.subRepo.ListSubAccounts(ctx, entID)
	if err != nil {
		return nil, err
	}

	responses := make([]*model.SubAccountResponse, len(subs))
	for i, sa := range subs {
		responses[i] = toSubAccountResponse(sa)
	}
	return responses, nil
}

func (s *subAccountService) DeleteSubAccount(ctx context.Context, subAccountID string) error {
	saID, err := uuid.Parse(subAccountID)
	if err != nil {
		return errors.New("invalid sub-account ID")
	}
	return s.subRepo.DeleteSubAccount(ctx, saID)
}

func (s *subAccountService) RequireLiveness(ctx context.Context, req *model.RequireLivenessRequest) error {
	saID, err := uuid.Parse(req.SubAccountID)
	if err != nil {
		return errors.New("invalid sub-account ID")
	}

	sa, err := s.subRepo.GetSubAccount(ctx, saID)
	if err != nil || sa == nil {
		return ErrSubAccountNotFound
	}

	sa.Status = model.SubAccountStatusPending
	sa.UpdatedAt = time.Now()

	return s.subRepo.UpdateSubAccount(ctx, sa)
}

func (s *subAccountService) CompleteLiveness(ctx context.Context, req *model.CompleteLivenessRequest) error {
	saID, err := uuid.Parse(req.SubAccountID)
	if err != nil {
		return errors.New("invalid sub-account ID")
	}

	sa, err := s.subRepo.GetSubAccount(ctx, saID)
	if err != nil || sa == nil {
		return ErrSubAccountNotFound
	}

	now := time.Now()
	sa.LastLivenessAt = &now
	sa.Status = model.SubAccountStatusActive
	sa.UpdatedAt = now

	return s.subRepo.UpdateSubAccount(ctx, sa)
}

func (s *subAccountService) CheckLivenessRequired(ctx context.Context, userID string, accessType string) (bool, error) {
	return accessType == "financial" || accessType == "sensitive_tax", nil
}

func toSubAccountResponse(sa *model.SubAccount) *model.SubAccountResponse {
	resp := &model.SubAccountResponse{
		ID:           sa.ID.String(),
		EnterpriseID: sa.EnterpriseID.String(),
		UserID:       sa.UserID.String(),
		Role:         string(sa.Role),
		Status:       string(sa.Status),
	}
	if sa.LastLivenessAt != nil {
		resp.LastLivenessAt = sa.LastLivenessAt.Format(time.RFC3339)
	}
	return resp
}
