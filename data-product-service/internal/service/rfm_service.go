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
		MonetaryScore:      m,
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
