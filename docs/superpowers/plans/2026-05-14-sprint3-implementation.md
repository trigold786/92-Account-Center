# Sprint 3 Implementation Plan — Data Product + Desensitization + VictoriaMetrics

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add RFM user profiling, dynamic PII masking, and VictoriaMetrics monitoring to the Account Center platform.

**Architecture:** data-product-service connects directly to PostgreSQL for RFM/dashboard/funnel queries. api-gateway gets a response-capture middleware for regex-based field desensitization. All 7 services expose `/metrics` in Prometheus exposition format. VictoriaMetrics container added to Docker Compose for metric scraping.

**Tech Stack:** Go 1.21, Gin, PostgreSQL (lib/pq), VictoriaMetrics (single-node), Prometheus scrape format (pure stdlib).

**Spec:** `docs/superpowers/specs/2026-05-14-sprint3-design.md`

---

## Task 1: data-product-service — Models and Repository

**Files:**
- Create: `data-product-service/internal/model/rfm.go`
- Create: `data-product-service/internal/repository/data_repository.go`

- [ ] **Step 1: Create the RFM model file**

```go
package model

type RFMScore struct {
	UserID            int64   `json:"user_id"`
	RecencyScore      int     `json:"recency_score"`
	FrequencyScore    int     `json:"frequency_score"`
	MonetaryScore     int     `json:"monetary_score"`
	RFMSegment        string  `json:"rfm_segment"`
	RFMSegmentCN      string  `json:"rfm_segment_cn"`
	LastSubscriptionAt string  `json:"last_subscription_at"`
	TotalSubscriptions int     `json:"total_subscriptions"`
	TotalSpent        float64 `json:"total_spent"`
}

type RFMBatchRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required"`
}

type DashboardOverview struct {
	TotalUsers           int                        `json:"total_users"`
	TotalSubscriptions   int                        `json:"total_subscriptions"`
	TotalCreditsEarned   float64                    `json:"total_credits_earned"`
	TotalCreditsConsumed float64                    `json:"total_credits_consumed"`
	BlacklistActive      int                        `json:"blacklist_entries_active"`
	RegistrationTrend    []DailyCount               `json:"registration_trend"`
	CreditFlow           map[string]float64         `json:"credit_flow"`
	RFMDistribution      map[string]int             `json:"rfm_distribution"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type FunnelStep struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type SubscriptionFunnel struct {
	Steps []FunnelStep `json:"steps"`
}

type SubscriptionStats struct {
	Freq      int
	Monetary  float64
	LastSubAt string
}

type UserTierCount struct {
	Tier  int
	Count int
}
```

- [ ] **Step 2: Create the data repository file**

```go
package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"

	"github.com/lib/pq"
)

type DataRepository interface {
	GetSubscriptionStats(ctx context.Context, userID int64) (*model.SubscriptionStats, error)
	GetAllSubscriptionStats(ctx context.Context) (map[int64]*model.SubscriptionStats, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetTotalSubscriptions(ctx context.Context) (int, error)
	GetTotalCreditsByTypes(ctx context.Context, types []string) (float64, error)
	GetActiveBlacklistCount(ctx context.Context) (int, error)
	GetRegistrationTrend(ctx context.Context, days int) ([]model.DailyCount, error)
	GetCreditFlow(ctx context.Context) (map[string]float64, error)
	GetUserTierCounts(ctx context.Context) ([]model.UserTierCount, error)
	GetDistinctSubscriberCount(ctx context.Context, minTier int) (int, error)
}

type dataRepository struct {
	db *sql.DB
}

func NewDataRepository(db *sql.DB) DataRepository {
	return &dataRepository{db: db}
}

func (r *dataRepository) GetSubscriptionStats(ctx context.Context, userID int64) (*model.SubscriptionStats, error) {
	stats := &model.SubscriptionStats{}
	query := `SELECT COUNT(*), COALESCE(SUM(price), 0), COALESCE(TO_CHAR(MAX(end_time), 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '') FROM subscriptions WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&stats.Freq, &stats.Monetary, &stats.LastSubAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &model.SubscriptionStats{Freq: 0, Monetary: 0, LastSubAt: ""}, nil
		}
		return nil, err
	}
	return stats, nil
}

func (r *dataRepository) GetAllSubscriptionStats(ctx context.Context) (map[int64]*model.SubscriptionStats, error) {
	query := `SELECT user_id, COUNT(*), COALESCE(SUM(price), 0), COALESCE(TO_CHAR(MAX(end_time), 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '') FROM subscriptions GROUP BY user_id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]*model.SubscriptionStats)
	for rows.Next() {
		var uid int64
		stats := &model.SubscriptionStats{}
		if err := rows.Scan(&uid, &stats.Freq, &stats.Monetary, &stats.LastSubAt); err != nil {
			return nil, err
		}
		result[uid] = stats
	}
	return result, rows.Err()
}

func (r *dataRepository) GetTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetTotalSubscriptions(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetTotalCreditsByTypes(ctx context.Context, types []string) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(amount), 0) FROM credit_transactions WHERE type = ANY($1) AND status IN ('AVAILABLE', 'CONSUMED')`
	err := r.db.QueryRowContext(ctx, query, pq.Array(types)).Scan(&total)
	return total, err
}

func (r *dataRepository) GetActiveBlacklistCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM blacklist_entries`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetRegistrationTrend(ctx context.Context, days int) ([]model.DailyCount, error) {
	query := `SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS d, COUNT(*) FROM users WHERE created_at >= NOW() - interval '30 days' GROUP BY d ORDER BY d DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.DailyCount
	for rows.Next() {
		var dc model.DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, err
		}
		result = append(result, dc)
	}
	return result, rows.Err()
}

func (r *dataRepository) GetCreditFlow(ctx context.Context) (map[string]float64, error) {
	query := `SELECT type, COALESCE(SUM(amount), 0) FROM credit_transactions WHERE status IN ('AVAILABLE', 'CONSUMED') GROUP BY type`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var t string
		var amt float64
		if err := rows.Scan(&t, &amt); err != nil {
			return nil, err
		}
		result[t] = amt
	}
	return result, rows.Err()
}

func (r *dataRepository) GetUserTierCounts(ctx context.Context) ([]model.UserTierCount, error) {
	query := `SELECT identity_tier, COUNT(*) FROM users GROUP BY identity_tier ORDER BY identity_tier`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UserTierCount
	for rows.Next() {
		var tc model.UserTierCount
		if err := rows.Scan(&tc.Tier, &tc.Count); err != nil {
			return nil, err
		}
		result = append(result, tc)
	}
	return result, rows.Err()
}

func (r *dataRepository) GetDistinctSubscriberCount(ctx context.Context, minTier int) (int, error) {
	var count int
	query := `SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'ACTIVE'`
	if minTier > 0 {
		query += ` AND tier_level >= $1`
		err := r.db.QueryRowContext(ctx, query, minTier).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
```

- [ ] **Step 3: Update go.mod with PostgreSQL driver**

Run in `data-product-service/`:
```
$env:GOWORK="off"; go get github.com/lib/pq
```

This adds the `github.com/lib/pq` dependency needed for `_ "github.com/lib/pq"` import in main.go.

- [ ] **Step 4: Verify compilation**

Run: `$env:GOWORK="off"; go build ./...` in `data-product-service/`
Expected: compiles without errors

---

## Task 2: data-product-service — RFM Service

**Files:**
- Create: `data-product-service/internal/service/rfm_service.go`
- Create: `data-product-service/internal/service/dashboard_service.go`

- [ ] **Step 1: Create the RFM service**

```go
package service

import (
	"context"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
)

type RFMService interface {
	GetRFM(ctx context.Context, userID int64) (*model.RFMScore, error)
	GetRFMBatch(ctx context.Context, userIDs []int64) ([]*model.RFMScore, error)
	GetRFMDistribution(ctx context.Context) (map[string]int, error)
}

type rfmService struct {
	dataRepo repository.DataRepository
}

func NewRFMService(dataRepo repository.DataRepository) RFMService {
	return &rfmService{dataRepo: dataRepo}
}

func (s *rfmService) computeRecency(daysSinceLastSub int) int {
	switch {
	case daysSinceLastSub <= 30:
		return 5
	case daysSinceLastSub <= 60:
		return 4
	case daysSinceLastSub <= 90:
		return 3
	case daysSinceLastSub <= 180:
		return 2
	default:
		return 1
	}
}

func (s *rfmService) computeFrequency(freq int) int {
	switch {
	case freq >= 10:
		return 5
	case freq >= 5:
		return 4
	case freq >= 3:
		return 3
	case freq >= 2:
		return 2
	default:
		return 1
	}
}

func (s *rfmService) computeMonetary(monetary float64) int {
	switch {
	case monetary >= 1000:
		return 5
	case monetary >= 500:
		return 4
	case monetary >= 200:
		return 3
	case monetary >= 100:
		return 2
	default:
		return 1
	}
}

type segment struct {
	Key    string
	NameCN string
}

func (s *rfmService) classifySegment(r, f, m int) segment {
	rHigh := r >= 4
	fHigh := f >= 4
	mHigh := m >= 4

	switch {
	case rHigh && fHigh && mHigh:
		return segment{"CHAMPION", "重要价值客户"}
	case rHigh && !fHigh && mHigh:
		return segment{"PROMISING", "重要发展客户"}
	case !rHigh && fHigh && mHigh:
		return segment{"LOYAL", "重要保持客户"}
	case !rHigh && !fHigh && mHigh:
		return segment{"AT_RISK", "重要挽留客户"}
	case rHigh && fHigh && !mHigh:
		return segment{"POTENTIAL_LOYAL", "一般价值客户"}
	case rHigh && !fHigh && !mHigh:
		return segment{"NEW", "一般发展客户"}
	case !rHigh && fHigh && !mHigh:
		return segment{"NEED_ATTENTION", "一般保持客户"}
	default:
		return segment{"ABOUT_TO_LOSE", "一般挽留客户"}
	}
}

func (s *rfmService) statsToRFM(userID int64, stats *model.SubscriptionStats) *model.RFMScore {
	r := 1
	if stats.LastSubAt != "" && stats.Freq > 0 {
		r = s.computeRecency(0)
	}
	f := s.computeFrequency(stats.Freq)
	m := s.computeMonetary(stats.Monetary)
	seg := s.classifySegment(r, f, m)

	return &model.RFMScore{
		UserID:             userID,
		RecencyScore:       r,
		FrequencyScore:     f,
		MonetaryScore:     m,
		RFMSegment:         seg.Key,
		RFMSegmentCN:       seg.NameCN,
		LastSubscriptionAt: stats.LastSubAt,
		TotalSubscriptions: stats.Freq,
		TotalSpent:         stats.Monetary,
	}
}

func (s *rfmService) GetRFM(ctx context.Context, userID int64) (*model.RFMScore, error) {
	stats, err := s.dataRepo.GetSubscriptionStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.statsToRFM(userID, stats), nil
}

func (s *rfmService) GetRFMBatch(ctx context.Context, userIDs []int64) ([]*model.RFMScore, error) {
	allStats, err := s.dataRepo.GetAllSubscriptionStats(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*model.RFMScore, 0, len(userIDs))
	for _, uid := range userIDs {
		stats, ok := allStats[uid]
		if !ok {
			stats = &model.SubscriptionStats{}
		}
		results = append(results, s.statsToRFM(uid, stats))
	}
	return results, nil
}

func (s *rfmService) GetRFMDistribution(ctx context.Context) (map[string]int, error) {
	allStats, err := s.dataRepo.GetAllSubscriptionStats(ctx)
	if err != nil {
		return nil, err
	}

	dist := map[string]int{
		"CHAMPION": 0, "PROMISING": 0, "LOYAL": 0, "AT_RISK": 0,
		"POTENTIAL_LOYAL": 0, "NEW": 0, "NEED_ATTENTION": 0, "ABOUT_TO_LOSE": 0,
	}

	totalUsers, err := s.dataRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}

	usersWithSubs := make(map[int64]bool)
	for uid := range allStats {
		usersWithSubs[uid] = true
	}

	for uid, stats := range allStats {
		rfm := s.statsToRFM(uid, stats)
		dist[rfm.RFMSegment]++
	}

	usersWithoutSubs := totalUsers - len(usersWithSubs)
	if usersWithoutSubs > 0 {
		seg := s.classifySegment(1, 1, 1)
		dist[seg.Key] += usersWithoutSubs
	}

	return dist, nil
}
```

- [ ] **Step 2: Create the dashboard service**

```go
package service

import (
	"context"
	"math"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
)

type DashboardService interface {
	GetOverview(ctx context.Context) (*model.DashboardOverview, error)
	GetSubscriptionFunnel(ctx context.Context) (*model.SubscriptionFunnel, error)
}

type dashboardService struct {
	dataRepo repository.DataRepository
	rfmSvc   RFMService
}

func NewDashboardService(dataRepo repository.DataRepository, rfmSvc RFMService) DashboardService {
	return &dashboardService{dataRepo: dataRepo, rfmSvc: rfmSvc}
}

func (s *dashboardService) GetOverview(ctx context.Context) (*model.DashboardOverview, error) {
	totalUsers, err := s.dataRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}

	totalSubs, err := s.dataRepo.GetTotalSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	earned, err := s.dataRepo.GetTotalCreditsByTypes(ctx, []string{"EARN_REFERRAL", "EARN_VERIFY", "REFUND_SUB"})
	if err != nil {
		return nil, err
	}

	consumed, err := s.dataRepo.GetTotalCreditsByTypes(ctx, []string{"CONSUME_SUB"})
	if err != nil {
		return nil, err
	}

	blCount, err := s.dataRepo.GetActiveBlacklistCount(ctx)
	if err != nil {
		return nil, err
	}

	trend, err := s.dataRepo.GetRegistrationTrend(ctx, 30)
	if err != nil {
		return nil, err
	}

	flow, err := s.dataRepo.GetCreditFlow(ctx)
	if err != nil {
		return nil, err
	}

	rfmDist, err := s.rfmSvc.GetRFMDistribution(ctx)
	if err != nil {
		return nil, err
	}

	return &model.DashboardOverview{
		TotalUsers:           totalUsers,
		TotalSubscriptions:   totalSubs,
		TotalCreditsEarned:   earned,
		TotalCreditsConsumed: consumed,
		BlacklistActive:      blCount,
		RegistrationTrend:    trend,
		CreditFlow:           flow,
		RFMDistribution:      rfmDist,
	}, nil
}

func (s *dashboardService) GetSubscriptionFunnel(ctx context.Context) (*model.SubscriptionFunnel, error) {
	totalUsers, err := s.dataRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}
	if totalUsers == 0 {
		totalUsers = 1
	}

	tierCounts, err := s.dataRepo.GetUserTierCounts(ctx)
	if err != nil {
		return nil, err
	}

	tierMap := make(map[int]int)
	for _, tc := range tierCounts {
		tierMap[tc.Tier] = tc.Count
	}

	l1Plus := 0
	for tier, count := range tierMap {
		if tier >= 1 {
			l1Plus += count
		}
	}

	l2Plus, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 2)
	if err != nil {
		return nil, err
	}

	l3Plus, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 3)
	if err != nil {
		return nil, err
	}

	l4, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 4)
	if err != nil {
		return nil, err
	}

	pct := func(n int) float64 {
		return math.Round(float64(n)*10000/float64(totalUsers)) / 100
	}

	return &model.SubscriptionFunnel{
		Steps: []model.FunnelStep{
			{Name: "注册用户", Count: totalUsers, Percentage: 100.0},
			{Name: "实名用户 (L1+)", Count: l1Plus, Percentage: pct(l1Plus)},
			{Name: "订阅用户 (L2+)", Count: l2Plus, Percentage: pct(l2Plus)},
			{Name: "高级订阅 (L3+)", Count: l3Plus, Percentage: pct(l3Plus)},
			{Name: "顶级订阅 (L4)", Count: l4, Percentage: pct(l4)},
		},
	}, nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `$env:GOWORK="off"; go build ./...` in `data-product-service/`
Expected: compiles without errors

---

## Task 3: data-product-service — Handlers and Main Wiring

**Files:**
- Create: `data-product-service/internal/handler/rfm_handler.go`
- Create: `data-product-service/internal/handler/dashboard_handler.go`
- Create: `data-product-service/internal/handler/funnel_handler.go`
- Modify: `data-product-service/cmd/main.go`
- Modify: `data-product-service/Dockerfile`
- Modify: `docker-compose.yml` (data-product-service environment)

- [ ] **Step 1: Create RFM handler**

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type RFMHandler struct {
	rfmSvc service.RFMService
}

func NewRFMHandler(rfmSvc service.RFMService) *RFMHandler {
	return &RFMHandler{rfmSvc: rfmSvc}
}

func (h *RFMHandler) GetRFM(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	rfm, err := h.rfmSvc.GetRFM(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": rfm})
}

func (h *RFMHandler) GetRFMBatch(c *gin.Context) {
	var req model.RFMBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	results, err := h.rfmSvc.GetRFMBatch(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": results})
}
```

- [ ] **Step 2: Create dashboard handler**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type DashboardHandler struct {
	dashSvc service.DashboardService
}

func NewDashboardHandler(dashSvc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: dashSvc}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.dashSvc.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": overview})
}
```

- [ ] **Step 3: Create funnel handler**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type FunnelHandler struct {
	dashSvc service.DashboardService
}

func NewFunnelHandler(dashSvc service.DashboardService) *FunnelHandler {
	return &FunnelHandler{dashSvc: dashSvc}
}

func (h *FunnelHandler) GetSubscriptionFunnel(c *gin.Context) {
	funnel, err := h.dashSvc.GetSubscriptionFunnel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": funnel})
}
```

- [ ] **Step 4: Rewrite cmd/main.go with full wiring**

Replace entire `data-product-service/cmd/main.go` with:

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/handler"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "account_center")

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	dataRepo := repository.NewDataRepository(db)
	rfmSvc := service.NewRFMService(dataRepo)
	dashSvc := service.NewDashboardService(dataRepo, rfmSvc)

	rfmHandler := handler.NewRFMHandler(rfmSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)
	funnelHandler := handler.NewFunnelHandler(dashSvc)

	r := gin.Default()

	dataGroup := r.Group("/api/v1/data")
	{
		dataGroup.GET("/rfm/:user_id", rfmHandler.GetRFM)
		dataGroup.POST("/rfm/batch", rfmHandler.GetRFMBatch)
		dataGroup.GET("/dashboard/overview", dashHandler.GetOverview)
		dataGroup.GET("/funnel/subscription", funnelHandler.GetSubscriptionFunnel)
	}

	r.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30314")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Data product service starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

- [ ] **Step 5: Update Dockerfile**

Replace `data-product-service/Dockerfile` entirely with:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /data-product-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget
COPY --from=builder /data-product-service /usr/local/bin/data-product-service
EXPOSE 30314
USER nobody
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30314/health || exit 1
CMD ["data-product-service"]
```

- [ ] **Step 6: Update docker-compose.yml — data-product-service environment**

In `docker-compose.yml`, replace the `data-product-service` service block. Add environment variables and depends_on:

```yaml
  data-product-service:
    build:
      context: ./data-product-service
      dockerfile: Dockerfile
    container_name: data-product-service
    ports:
      - "30314:30314"
    environment:
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: account_center
      PORT: "30314"
    networks:
      - app_network
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://127.0.0.1:30314/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M
    restart: always
```

- [ ] **Step 7: Verify compilation**

Run: `$env:GOWORK="off"; go build ./...` in `data-product-service/`
Expected: compiles without errors

- [ ] **Step 8: Rebuild Docker and test RFM endpoint**

Run:
```
docker compose build data-product-service
docker compose up -d data-product-service
```

Wait for healthy, then test:
```powershell
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/rfm/6" -Method Get
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/dashboard/overview" -Method Get
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/funnel/subscription" -Method Get
```

Expected: All 3 endpoints return `{"code":200, "message":"success", "data":{...}}`.

---

## Task 4: Dynamic Desensitization Middleware (api-gateway)

**Files:**
- Modify: `api-gateway/cmd/main.go`

- [ ] **Step 1: Add desensitization types and middleware to api-gateway/cmd/main.go**

Add the following types and functions to `api-gateway/cmd/main.go` **before** the `main()` function, after the existing `wrapWriter` function (around line 312):

```go
type responseCaptureWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseCaptureWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

var phoneRegex = regexp.MustCompile(`"phone_number"\s*:\s*"(\d{3})\d{4}(\d{4})"`)
var emailRegex = regexp.MustCompile(`"email"\s*:\s*"([a-zA-Z0-9])[a-zA-Z0-9._%+\-]*@([^"]+)"`)
var ipAddrRegex = regexp.MustCompile(`"ip_address"\s*:\s*"(\d{1,3}\.)\d{1,3}\.\d{1,3}(\.\d{1,3})"`)

func desensitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/internal/") {
			c.Next()
			return
		}

		captureWriter := &responseCaptureWriter{ResponseWriter: c.Writer}
		c.Writer = captureWriter

		c.Next()

		if c.Writer.Status() != http.StatusOK {
			return
		}

		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return
		}

		if len(captureWriter.body) > 1048576 {
			return
		}

		accountID, _ := c.Get("user_id")
		if accountIDStr, ok := accountID.(string); ok && strings.HasPrefix(accountIDStr, "admin_") {
			return
		}

		body := string(captureWriter.body)
		body = phoneRegex.ReplaceAllString(body, `"phone_number":"$1****$2"`)
		body = emailRegex.ReplaceAllString(body, `"email":"$1***@$2"`)
		body = ipAddrRegex.ReplaceAllString(body, `"ip_address":"$1*.*$2"`)

		if body != string(captureWriter.body) {
			c.Header("X-Desensitized", "true")
		}
	}
}
```

Also add `"regexp"` to the imports at the top of the file.

- [ ] **Step 2: Register desensitization middleware in main()**

In `api-gateway/cmd/main.go`, in the `main()` function, add the desensitization middleware **after** the existing middleware stack and **before** the route registrations (after the JWT auth middleware block, around line 219):

Add this line:
```go
	r.Use(desensitizeMiddleware())
```

This goes right after the JWT auth middleware block and before `r.Any("/api/v1/account/*path", ...)`.

- [ ] **Step 3: Verify compilation**

Run: `$env:GOWORK="off"; go build ./...` in `api-gateway/`
Expected: compiles without errors

- [ ] **Step 4: Rebuild Docker and test desensitization**

Run:
```
docker compose build api-gateway
docker compose up -d api-gateway
```

Test through gateway with JWT:
```powershell
$login = Invoke-RestMethod -Uri "http://localhost:30302/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"credential":"13800138004","password":"Test@123456"}'
$token = $login.access_token
$headers = @{Authorization = "Bearer $token"}

Invoke-RestMethod -Uri "http://localhost:30300/api/v1/account/6/tier" -Method Get -Headers $headers
```

Expected: Responses containing phone_number show masked values like `138****1234`. Response header includes `X-Desensitized: true` when masking is applied.

---

## Task 5: VictoriaMetrics /metrics Endpoint (All 7 Services)

**Files:**
- Create: `monitoring/promscrape.yml`
- Modify: `api-gateway/cmd/main.go`
- Modify: `account-service/cmd/main.go`
- Modify: `auth-service/cmd/main.go`
- Modify: `notification-service/cmd/main.go`
- Modify: `credit-service/cmd/main.go`
- Modify: `compliance-service/cmd/main.go`
- Modify: `data-product-service/cmd/main.go`
- Modify: `docker-compose.yml`

- [ ] **Step 1: Create monitoring/promscrape.yml**

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'account-center'
    static_configs:
      - targets:
          - 'api-gateway:30300'
          - 'account-service:30301'
          - 'auth-service:30302'
          - 'notification-service:30311'
          - 'credit-service:30312'
          - 'compliance-service:30313'
          - 'data-product-service:30314'
```

- [ ] **Step 2: Add /metrics to api-gateway/cmd/main.go**

In `api-gateway/cmd/main.go`, add a `/metrics` endpoint **before** the `/health` route registration. Add `fmt`, `bytes`, `runtime`, and `sync/atomic` to imports.

Add a global counter variable near the top of main.go (after imports):
```go
var (
	gatewayRequestCount uint64
)
```

Add the metrics endpoint route in main():
```go
	r.GET("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"api-gateway\"} %d\n", atomic.LoadUint64(&gatewayRequestCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"api-gateway\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})
```

Also wrap the existing middleware to count requests — in `requestIDMiddleware`, add `atomic.AddUint64(&gatewayRequestCount, 1)` at the start of the handler.

- [ ] **Step 3: Add /metrics to the other 6 services**

For each of the remaining 6 services (account-service, auth-service, notification-service, credit-service, compliance-service, data-product-service), add a `/metrics` endpoint in their `cmd/main.go` following the same pattern.

In each service, add near the top:
```go
var (
	requestCount uint64
)
```

Add the route before the `/health` route:
```go
	r.GET("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"<SERVICE_NAME>\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"<SERVICE_NAME>\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})
```

Replace `<SERVICE_NAME>` with the actual service name for each:
- `account-service` → `"account-service"`
- `auth-service` → `"auth-service"`
- `notification-service` → `"notification-service"`
- `credit-service` → `"credit-service"`
- `compliance-service` → `"compliance-service"`
- `data-product-service` → `"data-product-service"`

Required imports to add to each service's main.go: `"bytes"`, `"fmt"`, `"runtime"`, `"sync/atomic"`

- [ ] **Step 4: Add VictoriaMetrics service to docker-compose.yml**

Add the following service to `docker-compose.yml` after the `data-product-service` block:

```yaml
  victoriametrics:
    image: victoriametrics/victoria-metrics:latest
    container_name: victoriametrics
    ports:
      - "20010:8428"
    volumes:
      - vm_data:/victoria-metrics-data
      - ./monitoring/promscrape.yml:/etc/promscrape.yml
    command:
      - "--promscrape.config=/etc/promscrape.yml"
      - "--storageDataPath=/victoria-metrics-data"
      - "--retentionPeriod=30d"
    networks:
      - app_network
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M
    restart: always
```

Also add `vm_data` to the `volumes:` section at the bottom:

```yaml
volumes:
  postgres_data:
  redis_data:
  vm_data:
```

- [ ] **Step 5: Verify all services compile**

Run for each service directory:
```
$env:GOWORK="off"; go build ./...
```

Expected: All 7 services compile without errors.

- [ ] **Step 6: Full Docker rebuild and verification**

Run:
```
docker compose build
docker compose up -d
```

Wait for all containers healthy, then verify:

1. **Metrics endpoints** — for each service:
```powershell
$ports = @(30300, 30301, 30302, 30311, 30312, 30313, 30314)
foreach ($p in $ports) {
    $r = Invoke-RestMethod -Uri "http://localhost:$p/metrics" -Method Get
    Write-Host "Port $p : $($r.Substring(0, 80))..."
}
```
Expected: Each returns Prometheus format text with `http_requests_total` and `go_goroutines`.

2. **VictoriaMetrics scraping**:
```powershell
Invoke-RestMethod -Uri "http://localhost:20010/api/v1/targets" -Method Get
```
Expected: All 7 targets show health "UP".

3. **Query metrics**:
```powershell
Invoke-RestMethod -Uri "http://localhost:20010/api/v1/query?query=http_requests_total" -Method Get
```
Expected: Returns time-series data from all 7 services.

---

## Task 6: Full E2E Validation

**Files:** No new files.

- [ ] **Step 1: Health check all services**

```powershell
$services = @(
    @{name="api-gateway"; port=30300},
    @{name="account-service"; port=30301},
    @{name="auth-service"; port=30302},
    @{name="notification-service"; port=30311},
    @{name="credit-service"; port=30312},
    @{name="compliance-service"; port=30313},
    @{name="data-product-service"; port=30314}
)
foreach ($svc in $services) {
    $r = Invoke-RestMethod -Uri "http://localhost:$($svc.port)/health" -Method Get
    Write-Host "$($svc.name): OK"
}
```

Expected: All 7 return OK.

- [ ] **Step 2: Register + Login + JWT**

```powershell
$regBody = @{phone_number="13900139010"; account_id="sprint3_test"; password="Test@123456"; agree_to_terms=$true} | ConvertTo-Json
$reg = Invoke-RestMethod -Uri "http://localhost:30301/api/v1/account/register" -Method Post -ContentType "application/json" -Body $regBody
Write-Host "Register: user_id=$($reg.id)"

$loginBody = @{credential="13900139010"; password="Test@123456"} | ConvertTo-Json
$login = Invoke-RestMethod -Uri "http://localhost:30302/api/v1/auth/login" -Method Post -ContentType "application/json" -Body $loginBody
$token = $login.access_token
$userId = $login.user_id
Write-Host "Login: user_id=$userId token_length=$($token.Length)"
```

Expected: Registration returns user ID, login returns JWT token.

- [ ] **Step 3: Test RFM API**

```powershell
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/rfm/$userId" -Method Get
```

Expected: Returns RFM scores (all 1s for new user with no subscriptions).

- [ ] **Step 4: Test Dashboard Overview**

```powershell
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/dashboard/overview" -Method Get
```

Expected: Returns total_users, registration_trend, credit_flow, rfm_distribution.

- [ ] **Step 5: Test Subscription Funnel**

```powershell
Invoke-RestMethod -Uri "http://localhost:30314/api/v1/data/funnel/subscription" -Method Get
```

Expected: Returns 5-step funnel with percentages.

- [ ] **Step 6: Test desensitization through gateway**

```powershell
$headers = @{Authorization = "Bearer $token"}
Invoke-RestMethod -Uri "http://localhost:30300/api/v1/data/dashboard/overview" -Method Get -Headers $headers
```

Expected: Dashboard works through gateway (no PII in this response, but the middleware runs).

Also test a response that would have PII (if any endpoint returns phone_number/email through the gateway, verify it's masked).

- [ ] **Step 7: Test VictoriaMetrics scraping**

```powershell
Invoke-RestMethod -Uri "http://localhost:20010/api/v1/targets" -Method Get
Invoke-RestMethod -Uri "http://localhost:20010/api/v1/query?query=http_requests_total" -Method Get
```

Expected: All targets UP, metrics query returns data.

- [ ] **Step 8: Verify all containers healthy**

```powershell
docker compose ps
```

Expected: All 10 containers (7 services + postgres + redis + victoriametrics) show healthy/Up.

---

## Task 7: Commit Sprint 3

**Files:** All modified/created files.

- [ ] **Step 1: Stage all changes**

```bash
git add data-product-service/ api-gateway/cmd/main.go account-service/cmd/main.go auth-service/cmd/main.go notification-service/cmd/main.go credit-service/cmd/main.go compliance-service/cmd/main.go docker-compose.yml monitoring/ docs/superpowers/specs/2026-05-14-sprint3-design.md docs/superpowers/plans/2026-05-14-sprint3-implementation.md
```

- [ ] **Step 2: Create feature branch and commit**

```bash
git checkout -b feature/sprint3-data-product-desensitization-metrics
git commit -m "feat: Sprint 3 - RFM profiling, dynamic desensitization, VictoriaMetrics

- data-product-service: RFM 5-score 8-segment user profiling
- data-product-service: monitoring dashboard + subscription funnel APIs
- api-gateway: dynamic PII masking middleware (phone/email/IP)
- All 7 services: /metrics endpoint with Prometheus exposition format
- Docker Compose: VictoriaMetrics container on port 20010
- monitoring/promscrape.yml: scrape config for all services"
```

- [ ] **Step 3: Push feature branch**

```bash
git push -u origin feature/sprint3-data-product-desensitization-metrics
```
