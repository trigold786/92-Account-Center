package model

// ChangePasswordRequest represents the request body for password change.
type ChangePasswordRequest struct {
	CurrentPassword   string `json:"current_password,omitempty"`
	NewPassword      string `json:"new_password" validate:"required,min=8,max=20"`
	ConfirmPassword  string `json:"confirm_password" validate:"required"`
	VerificationCode string `json:"verification_code"`
	VerificationType string `json:"verification_type" validate:"required,oneof=sms_code email_otp password"`
}

// SendVerificationCodeRequest represents the request body for sending verification code.
type SendVerificationCodeRequest struct {
	ContactType  string `json:"contact_type" validate:"required,oneof=phone email"`
	ContactValue string `json:"contact_value" validate:"required"`
}