package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

func TestTrackAdEvent(t *testing.T) {
	svc := NewAdEventService(nil)

	e, err := svc.TrackAdEvent(context.Background(), &model.AdEvent{
		UserID:    1,
		EventType: "ad_splash_shown",
		AdID:      "ad-001",
		Placement: "home",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestGetAdMetrics(t *testing.T) {
	svc := NewAdEventService(nil)

	svc.TrackAdEvent(context.Background(), &model.AdEvent{UserID: 1, EventType: "ad_splash_shown", Placement: "home"})
	svc.TrackAdEvent(context.Background(), &model.AdEvent{UserID: 1, EventType: "ad_banner_shown", Placement: "home"})
	svc.TrackAdEvent(context.Background(), &model.AdEvent{UserID: 1, EventType: "ad_clicked", Placement: "home"})
	svc.TrackAdEvent(context.Background(), &model.AdEvent{UserID: 2, EventType: "ad_splash_shown", Placement: "profile"})
	svc.TrackAdEvent(context.Background(), &model.AdEvent{UserID: 2, EventType: "ad_clicked", Placement: "home", Converted: true})

	metrics, err := svc.GetAdMetrics(context.Background(), "", "home")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics["impressions"] != 2 {
		t.Fatalf("expected 2 impressions, got %d", metrics["impressions"])
	}
	if metrics["clicks"] != 2 {
		t.Fatalf("expected 2 clicks, got %d", metrics["clicks"])
	}
	if metrics["conversions"] != 1 {
		t.Fatalf("expected 1 conversion, got %d", metrics["conversions"])
	}
}
