package service

import (
	"context"
	"testing"
	"time"
)

func TestCalculateReminderDays(t *testing.T) {
	svc := NewRenewalService(nil, nil)
	now := time.Now()

	expiry7 := now.Add(7 * 24 * time.Hour)
	reminders := svc.GetDueReminders(context.Background(), expiry7)
	foundT7 := false
	for _, r := range reminders {
		if r.Type == "T-7" {
			foundT7 = true
			break
		}
	}
	if !foundT7 {
		t.Fatal("expected T-7 reminder for 7-day expiry")
	}

	expiryFar := now.Add(30 * 24 * time.Hour)
	remindersFar := svc.GetDueReminders(context.Background(), expiryFar)
	if len(remindersFar) > 0 {
		t.Fatal("expected no reminders for 30-day expiry")
	}
}

func TestSendReminderMultichannel(t *testing.T) {
	svc := NewRenewalService(nil, nil)
	err := svc.SendReminder(context.Background(), 1, "test@example.com", "push", "T-7")
	if err != nil {
		t.Fatalf("SendReminder failed: %v", err)
	}
}
