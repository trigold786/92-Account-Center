package model

const (
	QRCodeStatusPending   = "pending"
	QRCodeStatusScanned   = "scanned"
	QRCodeStatusConfirmed = "confirmed"
	QRCodeStatusExpired   = "expired"
	QRCodeTTL             = 300
)

type QRCodeGenerateResponse struct {
	CodeID    string `json:"code_id"`
	ExpiresIn int    `json:"expires_in"`
}

type QRCodeStatusResponse struct {
	CodeID string `json:"code_id"`
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
}

type QRCodeScanRequest struct {
	CodeID string `json:"code_id" binding:"required"`
	UserID int64  `json:"user_id" binding:"required"`
}

type QRCodeConfirmRequest struct {
	CodeID string `json:"code_id" binding:"required"`
	UserID int64  `json:"user_id" binding:"required"`
}
