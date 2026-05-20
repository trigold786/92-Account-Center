package service

import (
	"context"
	"time"
)

type RenewalReminder struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Channel   string `json:"channel"`
	Type      string `json:"type"`
	DaysUntil int    `json:"days_until"`
}

type RenewalService struct {
	notifClient interface{}
	userRepo    interface{}
}

func NewRenewalService(notifClient, userRepo interface{}) *RenewalService {
	return &RenewalService{notifClient: notifClient, userRepo: userRepo}
}

func (s *RenewalService) GetDueReminders(ctx context.Context, expiryDate time.Time) []RenewalReminder {
	now := time.Now()
	daysUntil := int(expiryDate.Sub(now).Hours() / 24)

	var reminders []RenewalReminder
	switch daysUntil {
	case 7:
		reminders = append(reminders, RenewalReminder{Type: "T-7", DaysUntil: 7, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-7", DaysUntil: 7, Channel: "email"})
	case 3:
		reminders = append(reminders, RenewalReminder{Type: "T-3", DaysUntil: 3, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-3", DaysUntil: 3, Channel: "sms"})
	case 1:
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "push"})
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "sms"})
		reminders = append(reminders, RenewalReminder{Type: "T-1", DaysUntil: 1, Channel: "email"})
	}
	return reminders
}

func (s *RenewalService) SendReminder(ctx context.Context, userID int64, email, channel, reminderType string) error {
	return nil
}
