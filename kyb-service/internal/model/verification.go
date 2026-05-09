package model

type EnterpriseInfoRequest struct {
	UserID                  string `json:"user_id" binding:"required"`
	CompanyName             string `json:"company_name" binding:"required"`
	UnifiedSocialCreditCode string `json:"unified_social_credit_code" binding:"required"`
	LegalPersonName         string `json:"legal_person_name" binding:"required"`
	LegalPersonIDNumber     string `json:"legal_person_id_number" binding:"required"`
	BankName                string `json:"bank_name" binding:"required"`
	BankAccountNumber       string `json:"bank_account_number" binding:"required"`
}

type MicroPaymentVerifyRequest struct {
	EnterpriseID string  `json:"enterprise_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
}

type FaceVerifyRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	Token       string `json:"token" binding:"required"`
}

type MicroPaymentInitResponse struct {
	EnterpriseID    string  `json:"enterprise_id"`
	Amount          float64 `json:"amount"`
	BankName        string  `json:"bank_name"`
	BankAccountMask string  `json:"bank_account_mask"`
}

type EnterpriseStatusResponse struct {
	EnterpriseID           string               `json:"enterprise_id"`
	CompanyName            string               `json:"company_name"`
	VerificationStatus     string               `json:"verification_status"`
	MicroPaymentStatus     string               `json:"micro_payment_status"`
	MicroPaymentAmount     float64              `json:"micro_payment_amount,omitempty"`
	FaceVerificationStatus string               `json:"face_verification_status"`
	FaceVerificationScore  float64              `json:"face_verification_score,omitempty"`
	CreatedAt              string               `json:"created_at"`
	UpdatedAt              string               `json:"updated_at"`
}

type EnterpriseSubmitResponse struct {
	EnterpriseID  string `json:"enterprise_id"`
	VerificationStatus string `json:"verification_status"`
	Message       string `json:"message"`
}