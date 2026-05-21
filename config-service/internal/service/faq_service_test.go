package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

func TestListFAQs(t *testing.T) {
	svc := NewFAQService(nil)

	svc.CreateFAQ(context.Background(), &model.FAQ{Category: "billing", Question: "Q1", SortOrder: 2})
	svc.CreateFAQ(context.Background(), &model.FAQ{Category: "billing", Question: "Q2", SortOrder: 1})
	svc.CreateFAQ(context.Background(), &model.FAQ{Category: "account", Question: "Q3", SortOrder: 0})

	all, err := svc.ListFAQs(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}

	billing, err := svc.ListFAQs(context.Background(), "billing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(billing) != 2 {
		t.Fatalf("expected 2 billing, got %d", len(billing))
	}
	if billing[0].SortOrder > billing[1].SortOrder {
		t.Fatal("expected sorted by sort_order")
	}
}

func TestSearchFAQs(t *testing.T) {
	svc := NewFAQService(nil)

	svc.CreateFAQ(context.Background(), &model.FAQ{Question: "How to pay?", Answer: "Use Alipay", Tags: "payment"})
	svc.CreateFAQ(context.Background(), &model.FAQ{Question: "Reset password", Answer: "Go to settings", Tags: "account"})

	results, err := svc.SearchFAQs(context.Background(), "pay")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Question != "How to pay?" {
		t.Fatalf("unexpected result: %s", results[0].Question)
	}
}

func TestCreateFAQ(t *testing.T) {
	svc := NewFAQService(nil)

	faq, err := svc.CreateFAQ(context.Background(), &model.FAQ{
		Category:  "general",
		Question:  "What is this?",
		Answer:    "A platform",
		SortOrder: 1,
		Tags:      "intro",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if faq.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}
