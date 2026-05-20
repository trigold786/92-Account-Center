package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

const (
	creditRate        = 0.01
	maxCreditDiscount = 0.50
)

func getPricingTiers() []model.PricingTier {
	return []model.PricingTier{
		{
			TierLevel:    2,
			Name:         "基础版",
			MonthlyPrice: 9.9,
			YearlyPrice:  99,
			Features: []model.TierFeature{
				{Name: "基础功能", Description: "包含核心账户管理功能", Included: true},
				{Name: "积分系统", Description: "标准积分倍率 1x", Included: true},
				{Name: "AI 调用", Description: "每月 100 次 AI 调用额度", Included: true},
				{Name: "专属客服", Description: "工作日在线客服支持", Included: false},
				{Name: "高级分析", Description: "高级数据分析功能", Included: false},
			},
			Entitlements: []model.TierEntitlement{
				{FeatureCode: "ai_call", Quota: 100, Unit: "次/月"},
				{FeatureCode: "credit_multiplier", Quota: 1, Unit: "倍"},
			},
		},
		{
			TierLevel:    3,
			Name:         "专业版",
			MonthlyPrice: 29.9,
			YearlyPrice:  299,
			Features: []model.TierFeature{
				{Name: "基础功能", Description: "包含核心账户管理功能", Included: true},
				{Name: "积分系统", Description: "积分倍率 2x", Included: true},
				{Name: "AI 调用", Description: "每月 500 次 AI 调用额度", Included: true},
				{Name: "专属客服", Description: "7x12 在线客服支持", Included: true},
				{Name: "高级分析", Description: "高级数据分析功能", Included: false},
			},
			Entitlements: []model.TierEntitlement{
				{FeatureCode: "ai_call", Quota: 500, Unit: "次/月"},
				{FeatureCode: "credit_multiplier", Quota: 2, Unit: "倍"},
			},
		},
		{
			TierLevel:    4,
			Name:         "企业版",
			MonthlyPrice: 99.9,
			YearlyPrice:  999,
			Features: []model.TierFeature{
				{Name: "基础功能", Description: "包含核心账户管理功能", Included: true},
				{Name: "积分系统", Description: "积分倍率 3x", Included: true},
				{Name: "AI 调用", Description: "每月 2000 次 AI 调用额度", Included: true},
				{Name: "专属客服", Description: "7x24 专属客服支持", Included: true},
				{Name: "高级分析", Description: "高级数据分析功能", Included: true},
			},
			Entitlements: []model.TierEntitlement{
				{FeatureCode: "ai_call", Quota: 2000, Unit: "次/月"},
				{FeatureCode: "credit_multiplier", Quota: 3, Unit: "倍"},
			},
		},
	}
}

type PricingHandler struct{}

func NewPricingHandler() *PricingHandler {
	return &PricingHandler{}
}

func (h *PricingHandler) GetPricing(c *gin.Context) {
	data := model.PricingPageData{
		Tiers:             getPricingTiers(),
		CreditRate:        creditRate,
		MaxCreditDiscount: maxCreditDiscount,
	}
	c.JSON(http.StatusOK, data)
}

func (h *PricingHandler) CalculateDiscount(c *gin.Context) {
	var req model.CreditDiscountCalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: price must be > 0 and credits_used must be >= 0"})
		return
	}

	creditValue := float64(req.CreditsUsed) * creditRate
	maxDiscountValue := req.Price * maxCreditDiscount

	if creditValue > maxDiscountValue {
		creditValue = maxDiscountValue
	}

	finalPrice := math.Round((req.Price-creditValue)*100) / 100
	if finalPrice < 0 {
		finalPrice = 0
	}

	discountPercent := math.Round((creditValue/req.Price)*10000) / 100

	c.JSON(http.StatusOK, model.CreditDiscountCalcResponse{
		OriginalPrice:   req.Price,
		CreditsUsed:     req.CreditsUsed,
		CreditValue:     math.Round(creditValue*100) / 100,
		FinalPrice:      finalPrice,
		DiscountPercent: discountPercent,
	})
}
