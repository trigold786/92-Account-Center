package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type mockGuestRepo struct {
	guests map[string]*model.GuestSession
}

func newMockGuestRepo() *mockGuestRepo {
	return &mockGuestRepo{guests: make(map[string]*model.GuestSession)}
}

func (m *mockGuestRepo) CreateGuest(_ context.Context, g *model.GuestSession) error {
	m.guests[g.AccountID] = g
	return nil
}

func (m *mockGuestRepo) GetByAccountID(_ context.Context, accountID string) (*model.GuestSession, error) {
	g, ok := m.guests[accountID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return g, nil
}

func (m *mockGuestRepo) UpdateGuest(_ context.Context, g *model.GuestSession) error {
	m.guests[g.AccountID] = g
	return nil
}

func TestCreateGuest(t *testing.T) {
	repo := newMockGuestRepo()
	svc := NewGuestService(repo)

	guest, err := svc.CreateGuest(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if guest.AccountID == "" {
		t.Fatal("expected non-empty account ID")
	}
	if guest.Status != "active" {
		t.Fatalf("expected active, got %s", guest.Status)
	}
	if guest.DeviceID != "device-123" {
		t.Fatalf("expected device-123, got %s", guest.DeviceID)
	}
}

func TestUpgradeGuest(t *testing.T) {
	repo := newMockGuestRepo()
	svc := NewGuestService(repo)

	guest, _ := svc.CreateGuest(context.Background(), "device-456")

	upgraded, err := svc.UpgradeGuest(context.Background(), &model.UpgradeGuestRequest{
		AccountID: guest.AccountID,
		Email:     "test@example.com",
		Phone:     "13800138000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upgraded.Status != "upgraded" {
		t.Fatalf("expected upgraded, got %s", upgraded.Status)
	}
	if upgraded.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", upgraded.Email)
	}
	if upgraded.Phone != "13800138000" {
		t.Fatalf("expected 13800138000, got %s", upgraded.Phone)
	}
}

func TestUpgradeGuestNotFound(t *testing.T) {
	repo := newMockGuestRepo()
	svc := NewGuestService(repo)

	_, err := svc.UpgradeGuest(context.Background(), &model.UpgradeGuestRequest{
		AccountID: "nonexistent",
		Email:     "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent guest")
	}
}
