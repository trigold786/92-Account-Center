package service

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type LeaderboardRepository interface {
	GetTopReferrers(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
	GetUserRank(ctx context.Context, userID int64) (int64, error)
	GetNearbyEntries(ctx context.Context, rank int64, offset int) ([]model.LeaderboardEntry, error)
}

type LeaderboardService struct {
	repo LeaderboardRepository
}

func NewLeaderboardService(repo LeaderboardRepository) *LeaderboardService {
	return &LeaderboardService{repo: repo}
}

func (s *LeaderboardService) GetTopReferrers(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if s.repo != nil {
		return s.repo.GetTopReferrers(ctx, limit)
	}
	entries := make([]model.LeaderboardEntry, 0, limit)
	for i := 1; i <= limit; i++ {
		entries = append(entries, model.LeaderboardEntry{
			Rank:        int64(i),
			UserID:      int64(100 + i),
			Score:       int64(500 - i*20),
			DisplayName: fmt.Sprintf("User%d", 100+i),
		})
	}
	return entries, nil
}

func (s *LeaderboardService) GetNearbyRank(ctx context.Context, userID int64, offset int) ([]model.LeaderboardEntry, error) {
	if offset <= 0 {
		offset = 5
	}
	if s.repo != nil {
		rank, err := s.repo.GetUserRank(ctx, userID)
		if err != nil {
			return nil, err
		}
		return s.repo.GetNearbyEntries(ctx, rank, offset)
	}
	var entries []model.LeaderboardEntry
	for i := -offset; i <= offset; i++ {
		rank := 10 + i
		if rank < 1 {
			continue
		}
		entries = append(entries, model.LeaderboardEntry{
			Rank:        int64(rank),
			UserID:      int64(200 + rank),
			Score:       int64(500 - rank*20),
			DisplayName: fmt.Sprintf("NearbyUser%d", rank),
		})
	}
	return entries, nil
}

func (s *LeaderboardService) GetSocialProof(ctx context.Context) ([]model.SocialProofMessage, error) {
	return []model.SocialProofMessage{
		{Message: "5 people upgraded today", Timestamp: "today"},
		{Message: "12 new users joined this hour", Timestamp: "this_hour"},
		{Message: "3 people near you subscribed to Pro", Timestamp: "recently"},
	}, nil
}
