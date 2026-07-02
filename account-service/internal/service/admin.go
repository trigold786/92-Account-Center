package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
)

var (
	ErrInvalidCreditType   = errors.New("invalid credit type")
	ErrCreditAdjustFailed  = errors.New("credit adjustment failed")
	ErrEnterpriseNotFound  = errors.New("enterprise not found")
)

type AdminService interface {
	ListUsers(ctx context.Context, req *model.AdminUserListRequest) (*model.AdminUserListResponse, error)
	GetUserDetail(ctx context.Context, userID int64) (*model.User, error)
	UpdateUserStatus(ctx context.Context, adminID string, userID int64, req *model.AdminStatusUpdateRequest) error
	AdjustIdentityTier(ctx context.Context, adminID string, userID int64, req *model.AdminTierUpdateRequest) error
	AdjustCredits(ctx context.Context, adminID string, userID int64, req *model.AdminCreditAdjustRequest) error
	GetAuditLog(ctx context.Context, userID int64, page, pageSize int) ([]model.AuditLogEntry, int, error)
	ListPendingKYC(ctx context.Context) ([]model.EnterpriseKYC, error)
	ReviewKYC(ctx context.Context, enterpriseID string, action string, reviewer string) error
}

type CreditClient interface {
	AdjustCredits(ctx context.Context, userID int64, amount int64, adjustType string) error
}

type httpCreditClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPCreditClient(baseURL string) CreditClient {
	return &httpCreditClient{
		baseURL:    baseURL,
		httpClient: circuitbreaker.WrapHTTPClient(&http.Client{Timeout: 5 * time.Second}, "credit-service"),
	}
}

func (c *httpCreditClient) AdjustCredits(ctx context.Context, userID int64, amount int64, adjustType string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"amount":  amount,
	})
	url := fmt.Sprintf("%s/internal/v1/credits/%s", c.baseURL, adjustType)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("credit service request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("credit service returned status %d", resp.StatusCode)
	}
	return nil
}

type adminService struct {
	repo         repository.AdminRepository
	creditClient CreditClient
}

func NewAdminService(repo repository.AdminRepository, creditClient CreditClient) AdminService {
	return &adminService{repo: repo, creditClient: creditClient}
}

func (s *adminService) ListUsers(ctx context.Context, req *model.AdminUserListRequest) (*model.AdminUserListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	users, total, err := s.repo.ListUsers(ctx, req.Page, req.PageSize, req.Search, req.Status, req.Tier)
	if err != nil {
		return nil, err
	}

	if users == nil {
		users = []model.User{}
	}

	return &model.AdminUserListResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *adminService) GetUserDetail(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.repo.GetUserDetail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *adminService) UpdateUserStatus(ctx context.Context, adminID string, userID int64, req *model.AdminStatusUpdateRequest) error {
	existing, err := s.repo.GetUserDetail(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	return s.repo.UpdateUserStatus(ctx, userID, req.Status, req.Reason, adminID)
}

func (s *adminService) AdjustIdentityTier(ctx context.Context, adminID string, userID int64, req *model.AdminTierUpdateRequest) error {
	existing, err := s.repo.GetUserDetail(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	return s.repo.AdjustIdentityTier(ctx, userID, req.Tier, req.Reason, adminID)
}

func (s *adminService) AdjustCredits(ctx context.Context, adminID string, userID int64, req *model.AdminCreditAdjustRequest) error {
	existing, err := s.repo.GetUserDetail(ctx, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	if s.creditClient != nil {
		if err := s.creditClient.AdjustCredits(ctx, userID, req.Amount, req.Type); err != nil {
			return ErrCreditAdjustFailed
		}
	}

	details, _ := json.Marshal(map[string]interface{}{
		"amount": req.Amount,
		"type":   req.Type,
		"reason": req.Reason,
	})
	return s.repo.InsertAuditLog(ctx, userID, "adjust_credits", string(details), adminID)
}

func (s *adminService) GetAuditLog(ctx context.Context, userID int64, page, pageSize int) ([]model.AuditLogEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entries, total, err := s.repo.GetAuditLog(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if entries == nil {
		entries = []model.AuditLogEntry{}
	}
	return entries, total, nil
}

func (s *adminService) ListPendingKYC(ctx context.Context) ([]model.EnterpriseKYC, error) {
	entries, err := s.repo.ListPendingEnterprises(ctx)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []model.EnterpriseKYC{}
	}
	return entries, nil
}

func (s *adminService) ReviewKYC(ctx context.Context, enterpriseID string, action string, reviewer string) error {
	status := "approved"
	if action == "reject" {
		status = "rejected"
	}

	if err := s.repo.UpdateEnterpriseStatus(ctx, enterpriseID, status, reviewer); err != nil {
		if isNoRows(err) {
			return ErrEnterpriseNotFound
		}
		return err
	}

	details, _ := json.Marshal(map[string]interface{}{
		"enterprise_id": enterpriseID,
		"action":        action,
		"status":        status,
		"reviewer":      reviewer,
	})
	return s.repo.InsertAuditLog(ctx, 0, "kyc_review", string(details), reviewer)
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
