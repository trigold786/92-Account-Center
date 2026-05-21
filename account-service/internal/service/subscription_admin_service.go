package service

import (
	"context"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type SubAdminRepository interface {
	CreatePlan(ctx context.Context, p *model.Plan) error
	ListPlans(ctx context.Context) ([]*model.Plan, error)
	GetPlan(ctx context.Context, id int64) (*model.Plan, error)
	UpdatePlan(ctx context.Context, p *model.Plan) error
	DeletePlan(ctx context.Context, id int64) error
	CreateCoupon(ctx context.Context, c *model.Coupon) error
	ListCoupons(ctx context.Context) ([]*model.Coupon, error)
	UpdateCoupon(ctx context.Context, c *model.Coupon) error
	DeleteCoupon(ctx context.Context, id int64) error
}

type SubscriptionAdminService struct {
	repo SubAdminRepository
}

func NewSubscriptionAdminService(repo SubAdminRepository) *SubscriptionAdminService {
	return &SubscriptionAdminService{repo: repo}
}

func (s *SubscriptionAdminService) CreatePlan(ctx context.Context, name, displayName string, price float64, interval string, features interface{}) (*model.Plan, error) {
	p := &model.Plan{Name: name, DisplayName: displayName, Price: price, Interval: interval, Active: true}
	if f, ok := features.(map[string]interface{}); ok {
		p.Features = f
	}
	if err := s.repo.CreatePlan(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *SubscriptionAdminService) ListPlans(ctx context.Context) ([]*model.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *SubscriptionAdminService) UpdatePlan(ctx context.Context, p *model.Plan) error {
	return s.repo.UpdatePlan(ctx, p)
}

func (s *SubscriptionAdminService) DeletePlan(ctx context.Context, id int64) error {
	return s.repo.DeletePlan(ctx, id)
}

func (s *SubscriptionAdminService) CreateCoupon(ctx context.Context, code, discountType string, discountValue float64, maxUses, maxPerUser int) (*model.Coupon, error) {
	c := &model.Coupon{Code: code, DiscountType: discountType, DiscountValue: discountValue, MaxUses: maxUses, MaxPerUser: maxPerUser, Active: true}
	if err := s.repo.CreateCoupon(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *SubscriptionAdminService) ListCoupons(ctx context.Context) ([]*model.Coupon, error) {
	return s.repo.ListCoupons(ctx)
}

func (s *SubscriptionAdminService) UpdateCoupon(ctx context.Context, c *model.Coupon) error {
	return s.repo.UpdateCoupon(ctx, c)
}

func (s *SubscriptionAdminService) DeleteCoupon(ctx context.Context, id int64) error {
	return s.repo.DeleteCoupon(ctx, id)
}
