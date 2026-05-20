package service

import (
	"context"
	"sort"
)

type DashboardCard struct {
	Type  string                 `json:"type"`
	Title string                 `json:"title"`
	Order int                    `json:"order"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

type DashboardResponse struct {
	Level string         `json:"level"`
	Cards []DashboardCard `json:"cards"`
}

type ConfigClient interface {
	GetDashboardLayout(ctx context.Context, level string) ([]DashboardCard, error)
}

type DashboardService struct {
	configClient ConfigClient
}

func NewDashboardService(configClient ConfigClient) *DashboardService {
	return &DashboardService{configClient: configClient}
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID int64, level string) (*DashboardResponse, error) {
	cards := s.getCardsForLevel(level)
	if s.configClient != nil {
		if dynamicCards, err := s.configClient.GetDashboardLayout(ctx, level); err == nil && len(dynamicCards) > 0 {
			cards = dynamicCards
		}
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].Order < cards[j].Order })
	return &DashboardResponse{Level: level, Cards: cards}, nil
}

func (s *DashboardService) getCardsForLevel(level string) []DashboardCard {
	if level == "L0" || level == "" {
		return []DashboardCard{
			{Type: "upgrade_guide", Title: "升级引导", Order: 1, Data: map[string]interface{}{"show_register": true}},
			{Type: "features", Title: "功能预览", Order: 2, Data: map[string]interface{}{"free_tier": true}},
		}
	}
	if level == "L1" {
		return []DashboardCard{
			{Type: "profile", Title: "个人信息", Order: 1},
			{Type: "points", Title: "积分概览", Order: 2, Data: map[string]interface{}{"total": 0}},
			{Type: "upgrade_guide", Title: "升级到专业版", Order: 3},
		}
	}
	if level == "L2" || level == "L3" {
		return []DashboardCard{
			{Type: "profile", Title: "个人信息", Order: 1},
			{Type: "credit_balance", Title: "积分余额", Order: 2},
			{Type: "subscription", Title: "订阅状态", Order: 3},
			{Type: "referral", Title: "推荐进度", Order: 4},
		}
	}
	return []DashboardCard{
		{Type: "profile", Title: "个人信息", Order: 1},
		{Type: "credit_balance", Title: "积分余额", Order: 2},
		{Type: "subscription", Title: "订阅状态", Order: 3},
		{Type: "referral", Title: "推荐进度", Order: 4},
		{Type: "team", Title: "团队管理", Order: 5},
		{Type: "admin_panel", Title: "管理面板", Order: 6},
		{Type: "enterprise_settings", Title: "企业设置", Order: 7},
	}
}
