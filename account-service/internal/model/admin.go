package model

import "time"

// AdminUserListRequest is the query parameters for listing users in the admin panel.
type AdminUserListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	Tier     *int   `form:"tier"`
}

// AdminUserListResponse is the paginated result of an admin user listing query.
type AdminUserListResponse struct {
	Users    []User `json:"users"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// AdminStatusUpdateRequest changes a user's account status with a required reason.
type AdminStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=active frozen banned"`
	Reason string `json:"reason" binding:"required"`
}

// AdminTierUpdateRequest changes a user's membership tier.
type AdminTierUpdateRequest struct {
	Tier   int    `json:"tier" binding:"oneof=0 1 2 3 4"`
	Reason string `json:"reason" binding:"required"`
}

// AdminCreditAdjustRequest adjusts a user's credit balance by earn or consume.
type AdminCreditAdjustRequest struct {
	Amount int64  `json:"amount" binding:"required"`
	Reason string `json:"reason" binding:"required"`
	Type   string `json:"type" binding:"required,oneof=earn consume"`
}

// AuditLogEntry records an administrative action on a user account.
type AuditLogEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Details   string    `json:"details" db:"details"`
	Operator  string    `json:"operator" db:"operator"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// EnterpriseKYC stores enterprise know-your-customer verification data.
type EnterpriseKYC struct {
	EnterpriseID            string    `json:"enterprise_id" db:"enterprise_id"`
	UserID                  string    `json:"user_id" db:"user_id"`
	CompanyName             string    `json:"company_name" db:"company_name"`
	UnifiedSocialCreditCode string    `json:"unified_social_credit_code" db:"unified_social_credit_code"`
	LegalPersonName         string    `json:"legal_person_name" db:"legal_person_name"`
	VerificationStatus      string    `json:"verification_status" db:"verification_status"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
}

// KYCReviewRequest represents an admin action to approve or reject a KYC submission.
type KYCReviewRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
}
