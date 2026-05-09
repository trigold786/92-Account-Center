package model

import (
	"time"

	"github.com/google/uuid"
)

type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusVerified VerificationStatus = "verified"
	VerificationStatusFailed   VerificationStatus = "failed"
)

type MicroPaymentStatus string

const (
	MicroPaymentStatusPending  MicroPaymentStatus = "pending"
	MicroPaymentStatusVerified MicroPaymentStatus = "verified"
	MicroPaymentStatusFailed   MicroPaymentStatus = "failed"
)

type FaceVerificationStatus string

const (
	FaceVerificationStatusPending  FaceVerificationStatus = "pending"
	FaceVerificationStatusVerified FaceVerificationStatus = "verified"
	FaceVerificationStatusFailed   FaceVerificationStatus = "failed"
)

type Enterprise struct {
	EnterpriseID              uuid.UUID               `json:"enterprise_id" db:"enterprise_id"`
	UserID                    uuid.UUID               `json:"user_id" db:"user_id"`
	CompanyName               string                  `json:"company_name" db:"company_name"`
	UnifiedSocialCreditCode   string                  `json:"unified_social_credit_code" db:"unified_social_credit_code"`
	LegalPersonName           string                  `json:"legal_person_name" db:"legal_person_name"`
	LegalPersonIDNumber       string                  `json:"legal_person_id_number" db:"legal_person_id_number"`
	BankName                  string                  `json:"bank_name" db:"bank_name"`
	BankAccountNumber         string                  `json:"bank_account_number" db:"bank_account_number"`
	VerificationStatus        VerificationStatus      `json:"verification_status" db:"verification_status"`
	MicroPaymentStatus        MicroPaymentStatus      `json:"micro_payment_status" db:"micro_payment_status"`
	MicroPaymentAmount        float64                 `json:"micro_payment_amount" db:"micro_payment_amount"`
	FaceVerificationStatus    FaceVerificationStatus  `json:"face_verification_status" db:"face_verification_status"`
	FaceVerificationScore     float64                 `json:"face_verification_score" db:"face_verification_score"`
	CreatedAt                 time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time               `json:"updated_at" db:"updated_at"`
}

func (Enterprise) TableName() string {
	return "enterprises"
}