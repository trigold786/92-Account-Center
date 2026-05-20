package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type SubscriptionAdminHandler struct {
	svc *service.SubscriptionAdminService
}

func NewSubscriptionAdminHandler(svc *service.SubscriptionAdminService) *SubscriptionAdminHandler {
	return &SubscriptionAdminHandler{svc: svc}
}

func (h *SubscriptionAdminHandler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string      `json:"name"`
		DisplayName string      `json:"display_name"`
		Price       float64     `json:"price"`
		Interval    string      `json:"interval"`
		Features    interface{} `json:"features"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), req.Name, req.DisplayName, req.Price, req.Interval, req.Features)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *SubscriptionAdminHandler) ListPlans(c *gin.Context) {
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *SubscriptionAdminHandler) DeletePlan(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *SubscriptionAdminHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		MaxUses       int     `json:"max_uses"`
		MaxPerUser    int     `json:"max_per_user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	coupon, err := h.svc.CreateCoupon(c.Request.Context(), req.Code, req.DiscountType, req.DiscountValue, req.MaxUses, req.MaxPerUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, coupon)
}

func (h *SubscriptionAdminHandler) ListCoupons(c *gin.Context) {
	coupons, err := h.svc.ListCoupons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, coupons)
}
