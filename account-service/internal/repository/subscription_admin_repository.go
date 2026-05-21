package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type SubscriptionAdminRepository struct {
	db *sql.DB
}

func NewSubscriptionAdminRepository(db *sql.DB) *SubscriptionAdminRepository {
	return &SubscriptionAdminRepository{db: db}
}

func (r *SubscriptionAdminRepository) CreatePlan(ctx context.Context, p *model.Plan) error {
	p.CreatedAt = time.Now()
	featuresJSON, _ := json.Marshal(p.Features)
	return r.db.QueryRowContext(ctx,
		`INSERT INTO subscription_plans (name, display_name, price, interval, features, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		p.Name, p.DisplayName, p.Price, p.Interval, featuresJSON, p.Active, p.CreatedAt,
	).Scan(&p.ID)
}

func (r *SubscriptionAdminRepository) ListPlans(ctx context.Context) ([]*model.Plan, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, display_name, price, interval, features, active, created_at
		 FROM subscription_plans ORDER BY price`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []*model.Plan
	for rows.Next() {
		p := &model.Plan{}
		var featuresJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Price, &p.Interval, &featuresJSON, &p.Active, &p.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(featuresJSON, &p.Features)
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *SubscriptionAdminRepository) GetPlan(ctx context.Context, id int64) (*model.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, display_name, price, interval, features, active, created_at
		 FROM subscription_plans WHERE id=$1`, id)
	p := &model.Plan{}
	var featuresJSON []byte
	if err := row.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Price, &p.Interval, &featuresJSON, &p.Active, &p.CreatedAt); err != nil {
		return nil, err
	}
	json.Unmarshal(featuresJSON, &p.Features)
	return p, nil
}

func (r *SubscriptionAdminRepository) UpdatePlan(ctx context.Context, p *model.Plan) error {
	featuresJSON, _ := json.Marshal(p.Features)
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscription_plans SET name=$1, display_name=$2, price=$3, interval=$4, features=$5, active=$6 WHERE id=$7`,
		p.Name, p.DisplayName, p.Price, p.Interval, featuresJSON, p.Active, p.ID)
	return err
}

func (r *SubscriptionAdminRepository) DeletePlan(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscription_plans WHERE id=$1`, id)
	return err
}

func (r *SubscriptionAdminRepository) CreateCoupon(ctx context.Context, c *model.Coupon) error {
	c.CreatedAt = time.Now()
	return r.db.QueryRowContext(ctx,
		`INSERT INTO coupons (code, discount_type, discount_value, max_uses, max_per_user, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		c.Code, c.DiscountType, c.DiscountValue, c.MaxUses, c.MaxPerUser, c.Active, c.CreatedAt,
	).Scan(&c.ID)
}

func (r *SubscriptionAdminRepository) ListCoupons(ctx context.Context) ([]*model.Coupon, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, discount_type, discount_value, max_uses, current_uses, max_per_user, active, created_at
		 FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var coupons []*model.Coupon
	for rows.Next() {
		c := &model.Coupon{}
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MaxUses, &c.CurrentUses, &c.MaxPerUser, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, nil
}

func (r *SubscriptionAdminRepository) UpdateCoupon(ctx context.Context, c *model.Coupon) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE coupons SET code=$1, discount_type=$2, discount_value=$3, max_uses=$4, max_per_user=$5, active=$6 WHERE id=$7`,
		c.Code, c.DiscountType, c.DiscountValue, c.MaxUses, c.MaxPerUser, c.Active, c.ID)
	return err
}

func (r *SubscriptionAdminRepository) DeleteCoupon(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id=$1`, id)
	return err
}
