package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

func TestPlanCRUD(t *testing.T) {
	repo := &mockSubAdminRepo{}
	svc := NewSubscriptionAdminService(repo)
	plan, err := svc.CreatePlan(context.Background(), "test_plan", "测试套餐", 29.9, "monthly", map[string]interface{}{"feature_x": true})
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}
	if plan.Name != "test_plan" {
		t.Fatalf("unexpected name: %s", plan.Name)
	}
	plans, err := svc.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans failed: %v", err)
	}
	if len(plans) == 0 {
		t.Fatal("expected non-empty plans")
	}
	err = svc.DeletePlan(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}
}

func TestCouponCRUD(t *testing.T) {
	repo := &mockSubAdminRepo{}
	svc := NewSubscriptionAdminService(repo)
	coupon, err := svc.CreateCoupon(context.Background(), "WELCOME10", "percentage", 10, 100, 1)
	if err != nil {
		t.Fatalf("CreateCoupon failed: %v", err)
	}
	if coupon.Code != "WELCOME10" {
		t.Fatalf("unexpected code: %s", coupon.Code)
	}
	coupons, err := svc.ListCoupons(context.Background())
	if err != nil {
		t.Fatalf("ListCoupons failed: %v", err)
	}
	if len(coupons) == 0 {
		t.Fatal("expected non-empty coupons")
	}
}

type mockSubAdminRepo struct{}

func (m *mockSubAdminRepo) CreatePlan(ctx context.Context, p *model.Plan) error { p.ID = 1; return nil }
func (m *mockSubAdminRepo) ListPlans(ctx context.Context) ([]*model.Plan, error) {
	return []*model.Plan{{ID: 1, Name: "basic", Price: 9.9}}, nil
}
func (m *mockSubAdminRepo) GetPlan(ctx context.Context, id int64) (*model.Plan, error) {
	return &model.Plan{ID: id, Name: "pro", Price: 29.9}, nil
}
func (m *mockSubAdminRepo) UpdatePlan(ctx context.Context, p *model.Plan) error { return nil }
func (m *mockSubAdminRepo) DeletePlan(ctx context.Context, id int64) error { return nil }
func (m *mockSubAdminRepo) CreateCoupon(ctx context.Context, c *model.Coupon) error { c.ID = 1; return nil }
func (m *mockSubAdminRepo) ListCoupons(ctx context.Context) ([]*model.Coupon, error) {
	return []*model.Coupon{{ID: 1, Code: "WELCOME10", DiscountType: "percentage", DiscountValue: 10}}, nil
}
func (m *mockSubAdminRepo) UpdateCoupon(ctx context.Context, c *model.Coupon) error { return nil }
func (m *mockSubAdminRepo) DeleteCoupon(ctx context.Context, id int64) error { return nil }
