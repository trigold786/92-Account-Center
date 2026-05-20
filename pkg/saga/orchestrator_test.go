package saga

import (
	"context"
	"errors"
	"testing"
)

func TestSagaSuccess(t *testing.T) {
	var executionOrder []string
	saga := New("subscribe_flow", nil)
	saga.AddStep(NewStep("deduct_credits", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "deduct")
		return nil
	}, nil))
	saga.AddStep(NewStep("activate_subscription", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "activate")
		return nil
	}, nil))
	saga.AddStep(NewStep("grant_benefits", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "grant")
		return nil
	}, nil))

	err := saga.Execute(context.Background())
	if err != nil {
		t.Fatalf("Saga execution failed: %v", err)
	}
	if len(executionOrder) != 3 {
		t.Fatalf("expected 3 steps executed, got %d", len(executionOrder))
	}
	if saga.Status != StatusCompleted {
		t.Fatalf("expected completed status, got %v", saga.Status)
	}
}

func TestSagaCompensation(t *testing.T) {
	var compensationOrder []string
	saga := New("failed_flow", nil)
	saga.AddStep(NewStep("deduct", func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "deduct_comp"); return nil }))
	saga.AddStep(NewStep("activate", func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "activate_comp"); return nil }))
	saga.AddStep(NewStep("grant", func(ctx context.Context) error { return errors.New("grant failed") },
		func(ctx context.Context) error { compensationOrder = append(compensationOrder, "grant_comp"); return nil }))

	err := saga.Execute(context.Background())
	if err == nil {
		t.Fatal("expected saga execution error")
	}
	if len(compensationOrder) != 2 {
		t.Fatalf("expected 2 compensations, got %d: %v", len(compensationOrder), compensationOrder)
	}
	if saga.Status != StatusCompensated {
		t.Fatalf("expected compensated status, got %v", saga.Status)
	}
}

func TestIdempotencyKey(t *testing.T) {
	saga := New("idempotent_flow", nil)
	saga.SetID("unique_key_123")
	if saga.ID != "unique_key_123" {
		t.Fatalf("unexpected ID: %s", saga.ID)
	}
}
