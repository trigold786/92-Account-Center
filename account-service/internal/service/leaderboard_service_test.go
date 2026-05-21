package service

import (
	"context"
	"testing"
)

func TestGetTopReferrers(t *testing.T) {
	svc := NewLeaderboardService(nil)

	entries, err := svc.GetTopReferrers(context.Background(), 20)
	if err != nil {
		t.Fatalf("GetTopReferrers failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	if entries[0].Rank != 1 {
		t.Fatalf("expected first entry rank 1, got %d", entries[0].Rank)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Rank <= entries[i-1].Rank {
			t.Fatalf("entries not sorted by rank at index %d", i)
		}
	}
}

func TestGetNearbyRank(t *testing.T) {
	svc := NewLeaderboardService(nil)

	entries, err := svc.GetNearbyRank(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("GetNearbyRank failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry near rank")
	}
}

func TestGetSocialProof(t *testing.T) {
	svc := NewLeaderboardService(nil)

	messages, err := svc.GetSocialProof(context.Background())
	if err != nil {
		t.Fatalf("GetSocialProof failed: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected at least 1 social proof message")
	}
	for _, msg := range messages {
		if msg.Message == "" {
			t.Fatal("social proof message should not be empty")
		}
	}
}
