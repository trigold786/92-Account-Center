package model

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type SendMagicLinkRequest struct {
	Email     string `json:"email" binding:"required,email"`
	TargetURL string `json:"target_url" binding:"required,url"`
}

type VerifyMagicLinkRequest struct {
	Token string `json:"token" binding:"required"`
}

type SendEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type OTPResponse struct {
	ExpiresIn int `json:"expires_in"`
}

type MagicLinkResponse struct {
	MagicLink string `json:"magic_link"`
	ExpiresIn int    `json:"expires_in"`
}

type VerifyMagicLinkResponse struct {
	Email string `json:"email"`
}
