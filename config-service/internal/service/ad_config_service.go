package service

import (
	"context"
	"sync"
	"time"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

var noAdLevels = map[string]bool{"L2": true, "L3": true, "L4": true}

type AdConfigService struct {
	mu   sync.Mutex
	freq map[string]int
}

func NewAdConfigService(repo interface{}) *AdConfigService {
	return &AdConfigService{freq: make(map[string]int)}
}

func (s *AdConfigService) GetAdConfig(ctx context.Context, level string) (*model.AdConfig, error) {
	showAds := !noAdLevels[level]
	return &model.AdConfig{
		Placement:           "splash",
		PrimaryProvider:     "csj",
		BackupProvider:      "ylh",
		ShowAds:             showAds,
		VideoMaxDurationSec: 5,
		FrequencyPerHour:    3,
		EnabledLevels:       []string{"L0", "L1"},
	}, nil
}

func (s *AdConfigService) CheckFrequencyControl(ctx context.Context, userKey, placement string) (bool, error) {
	key := userKey + "_" + placement
	s.mu.Lock()
	defer s.mu.Unlock()
	count := s.freq[key]
	if count >= 3 {
		return false, nil
	}
	s.freq[key] = count + 1
	go func() {
		time.Sleep(time.Hour)
		s.mu.Lock()
		s.freq[key]--
		if s.freq[key] <= 0 {
			delete(s.freq, key)
		}
		s.mu.Unlock()
	}()
	return true, nil
}

func (s *AdConfigService) GetPrimarySDK(ctx context.Context) (string, error) {
	return "csj", nil
}

func (s *AdConfigService) GetBackupSDK(ctx context.Context) (string, error) {
	return "ylh", nil
}
