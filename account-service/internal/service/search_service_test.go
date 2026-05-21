package service

import (
	"context"
	"testing"
)

func TestSearch(t *testing.T) {
	svc := NewSearchService(nil)

	resp, err := svc.Search(context.Background(), "test", "", 1, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Query != "test" {
		t.Fatalf("expected query 'test', got %s", resp.Query)
	}
	if resp.Page != 1 {
		t.Fatalf("expected page 1, got %d", resp.Page)
	}
}

func TestSearchByType(t *testing.T) {
	svc := NewSearchService(nil)

	resp, err := svc.Search(context.Background(), "user", "users", 1, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Query != "user" {
		t.Fatalf("expected query 'user', got %s", resp.Query)
	}
}

func TestQuickActions(t *testing.T) {
	svc := NewSearchService(nil)

	resp, err := svc.QuickActions(context.Background(), 0)
	if err != nil {
		t.Fatalf("QuickActions failed: %v", err)
	}
	if resp.Tier != 0 {
		t.Fatalf("expected tier 0, got %d", resp.Tier)
	}
	if len(resp.Actions) == 0 {
		t.Fatal("expected at least 1 quick action for tier 0")
	}

	resp2, err := svc.QuickActions(context.Background(), 2)
	if err != nil {
		t.Fatalf("QuickActions failed: %v", err)
	}
	if resp2.Tier != 2 {
		t.Fatalf("expected tier 2, got %d", resp2.Tier)
	}
	if len(resp2.Actions) == 0 {
		t.Fatal("expected at least 1 quick action for tier 2")
	}
}
