package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/kyb-service/internal/model"
)

type EnterpriseRepository interface {
	Create(ctx context.Context, enterprise *model.Enterprise) error
	GetByID(ctx context.Context, enterpriseID uuid.UUID) (*model.Enterprise, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Enterprise, error)
	Update(ctx context.Context, enterprise *model.Enterprise) error
}

type enterpriseRepository struct {
	db *sql.DB
}

func NewEnterpriseRepository(db *sql.DB) EnterpriseRepository {
	return &enterpriseRepository{db: db}
}

func (r *enterpriseRepository) Create(ctx context.Context, enterprise *model.Enterprise) error {
	query := `
		INSERT INTO enterprises (
			enterprise_id, user_id, company_name, unified_social_credit_code,
			legal_person_name, legal_person_id_number, bank_name, bank_account_number,
			verification_status, micro_payment_status, micro_payment_amount,
			face_verification_status, face_verification_score, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		enterprise.EnterpriseID, enterprise.UserID, enterprise.CompanyName,
		enterprise.UnifiedSocialCreditCode, enterprise.LegalPersonName,
		enterprise.LegalPersonIDNumber, enterprise.BankName, enterprise.BankAccountNumber,
		enterprise.VerificationStatus, enterprise.MicroPaymentStatus,
		enterprise.MicroPaymentAmount, enterprise.FaceVerificationStatus,
		enterprise.FaceVerificationScore, enterprise.CreatedAt, enterprise.UpdatedAt,
	)
	return err
}

func (r *enterpriseRepository) GetByID(ctx context.Context, enterpriseID uuid.UUID) (*model.Enterprise, error) {
	query := `
		SELECT enterprise_id, user_id, company_name, unified_social_credit_code,
			legal_person_name, legal_person_id_number, bank_name, bank_account_number,
			verification_status, micro_payment_status, micro_payment_amount,
			face_verification_status, face_verification_score, created_at, updated_at
		FROM enterprises WHERE enterprise_id = $1
	`
	enterprise := &model.Enterprise{}
	err := r.db.QueryRowContext(ctx, query, enterpriseID).Scan(
		&enterprise.EnterpriseID, &enterprise.UserID, &enterprise.CompanyName,
		&enterprise.UnifiedSocialCreditCode, &enterprise.LegalPersonName,
		&enterprise.LegalPersonIDNumber, &enterprise.BankName, &enterprise.BankAccountNumber,
		&enterprise.VerificationStatus, &enterprise.MicroPaymentStatus,
		&enterprise.MicroPaymentAmount, &enterprise.FaceVerificationStatus,
		&enterprise.FaceVerificationScore, &enterprise.CreatedAt, &enterprise.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return enterprise, nil
}

func (r *enterpriseRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Enterprise, error) {
	query := `
		SELECT enterprise_id, user_id, company_name, unified_social_credit_code,
			legal_person_name, legal_person_id_number, bank_name, bank_account_number,
			verification_status, micro_payment_status, micro_payment_amount,
			face_verification_status, face_verification_score, created_at, updated_at
		FROM enterprises WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enterprises []*model.Enterprise
	for rows.Next() {
		enterprise := &model.Enterprise{}
		err := rows.Scan(
			&enterprise.EnterpriseID, &enterprise.UserID, &enterprise.CompanyName,
			&enterprise.UnifiedSocialCreditCode, &enterprise.LegalPersonName,
			&enterprise.LegalPersonIDNumber, &enterprise.BankName, &enterprise.BankAccountNumber,
			&enterprise.VerificationStatus, &enterprise.MicroPaymentStatus,
			&enterprise.MicroPaymentAmount, &enterprise.FaceVerificationStatus,
			&enterprise.FaceVerificationScore, &enterprise.CreatedAt, &enterprise.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		enterprises = append(enterprises, enterprise)
	}
	return enterprises, rows.Err()
}

func (r *enterpriseRepository) Update(ctx context.Context, enterprise *model.Enterprise) error {
	query := `
		UPDATE enterprises SET
			company_name = $1, unified_social_credit_code = $2,
			legal_person_name = $3, legal_person_id_number = $4,
			bank_name = $5, bank_account_number = $6,
			verification_status = $7, micro_payment_status = $8,
			micro_payment_amount = $9, face_verification_status = $10,
			face_verification_score = $11, updated_at = $12
		WHERE enterprise_id = $13
	`
	result, err := r.db.ExecContext(ctx, query,
		enterprise.CompanyName, enterprise.UnifiedSocialCreditCode,
		enterprise.LegalPersonName, enterprise.LegalPersonIDNumber,
		enterprise.BankName, enterprise.BankAccountNumber,
		enterprise.VerificationStatus, enterprise.MicroPaymentStatus,
		enterprise.MicroPaymentAmount, enterprise.FaceVerificationStatus,
		enterprise.FaceVerificationScore, enterprise.UpdatedAt,
		enterprise.EnterpriseID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("enterprise not found")
	}
	return nil
}