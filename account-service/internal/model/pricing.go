package model

type PricingTier struct {
	TierLevel    int               `json:"tier_level"`
	Name         string            `json:"name"`
	MonthlyPrice float64           `json:"monthly_price"`
	YearlyPrice  float64           `json:"yearly_price"`
	Features     []TierFeature     `json:"features"`
	Entitlements []TierEntitlement `json:"entitlements"`
}

type TierFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Included    bool   `json:"included"`
}

type TierEntitlement struct {
	FeatureCode string `json:"feature_code"`
	Quota       int    `json:"quota"`
	Unit        string `json:"unit"`
}

type PricingPageData struct {
	Tiers             []PricingTier `json:"tiers"`
	CreditRate        float64       `json:"credit_rate"`
	MaxCreditDiscount float64       `json:"max_credit_discount"`
}

type CreditDiscountCalcRequest struct {
	Price       float64 `json:"price" binding:"required,gt=0"`
	CreditsUsed int64   `json:"credits_used" binding:"gte=0"`
}

type CreditDiscountCalcResponse struct {
	OriginalPrice   float64 `json:"original_price"`
	CreditsUsed     int64   `json:"credits_used"`
	CreditValue     float64 `json:"credit_value"`
	FinalPrice      float64 `json:"final_price"`
	DiscountPercent float64 `json:"discount_percent"`
}
