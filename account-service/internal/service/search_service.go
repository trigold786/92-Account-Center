package service

import (
	"context"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type SearchRepository interface {
	SearchUsers(ctx context.Context, keyword string, page, size int) ([]model.SearchResult, int, error)
	SearchSubscriptions(ctx context.Context, keyword string, page, size int) ([]model.SearchResult, int, error)
	SearchOrders(ctx context.Context, keyword string, page, size int) ([]model.SearchResult, int, error)
}

type SearchService struct {
	repo SearchRepository
}

func NewSearchService(repo SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(ctx context.Context, keyword, searchType string, page, size int) (*model.SearchResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	var results []model.SearchResult
	var total int

	if s.repo == nil {
		results = s.mockSearch(keyword, searchType)
		total = len(results)
		start := (page - 1) * size
		if start > len(results) {
			start = len(results)
		}
		end := start + size
		if end > len(results) {
			end = len(results)
		}
		results = results[start:end]
	} else {
		var err error
		switch searchType {
		case "users":
			results, total, err = s.repo.SearchUsers(ctx, keyword, page, size)
		case "subscriptions":
			results, total, err = s.repo.SearchSubscriptions(ctx, keyword, page, size)
		case "orders":
			results, total, err = s.repo.SearchOrders(ctx, keyword, page, size)
		default:
			var allResults []model.SearchResult
			userResults, userTotal, _ := s.repo.SearchUsers(ctx, keyword, page, size)
			subResults, subTotal, _ := s.repo.SearchSubscriptions(ctx, keyword, page, size)
			orderResults, orderTotal, _ := s.repo.SearchOrders(ctx, keyword, page, size)
			allResults = append(allResults, userResults...)
			allResults = append(allResults, subResults...)
			allResults = append(allResults, orderResults...)
			total = userTotal + subTotal + orderTotal
			results = allResults
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return &model.SearchResponse{
		Query:   keyword,
		Results: results,
		Total:   total,
		Page:    page,
		Size:    size,
	}, nil
}

func (s *SearchService) mockSearch(keyword, searchType string) []model.SearchResult {
	var results []model.SearchResult
	if searchType == "" || searchType == "users" {
		results = append(results, model.SearchResult{
			Type:    "users",
			ID:      "1",
			Title:   "User: " + keyword,
			Summary: "Matching user account",
			Meta:    map[string]interface{}{"phone": "138****0001"},
		})
	}
	if searchType == "" || searchType == "subscriptions" {
		results = append(results, model.SearchResult{
			Type:    "subscriptions",
			ID:      "1",
			Title:   "Subscription: " + keyword,
			Summary: "Matching subscription",
			Meta:    map[string]interface{}{"plan": "pro"},
		})
	}
	if searchType == "" || searchType == "orders" {
		results = append(results, model.SearchResult{
			Type:    "orders",
			ID:      "1",
			Title:   "Order: " + keyword,
			Summary: "Matching order",
			Meta:    map[string]interface{}{"status": "completed"},
		})
	}
	return results
}

func (s *SearchService) QuickActions(ctx context.Context, tier int) (*model.QuickActionsResponse, error) {
	actions := s.getActionsForTier(tier)
	return &model.QuickActionsResponse{
		Tier:    tier,
		Actions: actions,
	}, nil
}

func (s *SearchService) getActionsForTier(tier int) []model.QuickAction {
	switch {
	case tier == 0:
		return []model.QuickAction{
			{ID: "upgrade", Label: "升级账户", Icon: "arrow-up", Description: "升级到基础版"},
			{ID: "profile", Label: "完善资料", Icon: "user", Description: "完善个人资料"},
		}
	case tier == 1:
		return []model.QuickAction{
			{ID: "upgrade", Label: "升级到专业版", Icon: "arrow-up", Description: "获取更多功能"},
			{ID: "credits", Label: "查看积分", Icon: "coins", Description: "查看积分余额"},
			{ID: "invite", Label: "邀请好友", Icon: "share", Description: "邀请好友获取奖励"},
		}
	default:
		return []model.QuickAction{
			{ID: "dashboard", Label: "数据看板", Icon: "chart", Description: "查看数据概览"},
			{ID: "team", Label: "团队管理", Icon: "users", Description: "管理团队成员"},
			{ID: "api", Label: "API管理", Icon: "code", Description: "管理API密钥"},
			{ID: "support", Label: "专属客服", Icon: "headset", Description: "联系专属客服"},
		}
	}
}
