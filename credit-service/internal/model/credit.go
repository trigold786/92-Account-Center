package model

type CreditAccount struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type CreditTransaction struct {
	ID              int64   `json:"id"`
	CreditAccountID int64   `json:"credit_account_id"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	ReferenceID     string  `json:"reference_id,omitempty"`
	Details         string  `json:"details,omitempty"`
	SM3Hash         string  `json:"sm3_hash"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
}

type EarnRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required"`
	ReferenceID string  `json:"reference_id"`
	Details     string  `json:"details"`
}

type ConsumeRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceID string  `json:"reference_id" binding:"required"`
	Details     string  `json:"details"`
}

type RefundRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ReferenceID string  `json:"reference_id" binding:"required"`
	Details     string  `json:"details"`
}

type CalculateDiscountRequest struct {
	UserID            string  `json:"user_id" binding:"required"`
	SubscriptionPrice float64 `json:"subscription_price" binding:"required,gt=0"`
}

type CalculateDiscountResponse struct {
	AvailableBalance float64 `json:"available_balance"`
	MaxDiscount      float64 `json:"max_discount"`
	RemainingToPay   float64 `json:"remaining_to_pay"`
}

type AccountResponse struct {
	UserID  int64   `json:"user_id"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

type TransactionListResponse struct {
	Transactions []CreditTransaction `json:"transactions"`
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
}

type RebateConfig struct {
	ID                  int64   `json:"id" db:"id"`
	SubscriptionCountMin int     `json:"subscription_count_min" db:"subscription_count_min"`
	SubscriptionCountMax int     `json:"subscription_count_max" db:"subscription_count_max"`
	RebatePercentage    float64 `json:"rebate_percentage" db:"rebate_percentage"`
	Description         string  `json:"description" db:"description"`
}

type ProcessSubscriptionPaidEvent struct {
	RefereeID         int64   `json:"referee_id"`
	SubscriptionPrice float64 `json:"subscription_price"`
	OrderID           string  `json:"order_id"`
}
