package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

func TestCreateExperiment(t *testing.T) {
	svc := NewABTestService()

	exp, err := svc.CreateExperiment(context.Background(), "button_color", []model.ABVariant{
		{Name: "control", Weight: 0.5},
		{Name: "blue", Weight: 0.5},
	})
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}
	if exp.Name != "button_color" {
		t.Fatalf("expected name 'button_color', got %s", exp.Name)
	}
	if len(exp.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(exp.Variants))
	}
	if exp.Status != "active" {
		t.Fatalf("expected status 'active', got %s", exp.Status)
	}
}

func TestDeterministicAssignment(t *testing.T) {
	svc := NewABTestService()

	exp, _ := svc.CreateExperiment(context.Background(), "test_exp", []model.ABVariant{
		{Name: "a", Weight: 0.5},
		{Name: "b", Weight: 0.5},
	})

	assign1, err := svc.AssignVariant(context.Background(), exp.ID, "user_123")
	if err != nil {
		t.Fatalf("AssignVariant failed: %v", err)
	}
	if assign1.Variant == "" {
		t.Fatal("expected non-empty variant")
	}

	assign2, err := svc.AssignVariant(context.Background(), exp.ID, "user_123")
	if err != nil {
		t.Fatalf("AssignVariant failed: %v", err)
	}
	if assign1.Variant != assign2.Variant {
		t.Fatalf("expected deterministic assignment, got %s then %s", assign1.Variant, assign2.Variant)
	}

	assign3, err := svc.AssignVariant(context.Background(), exp.ID, "user_456")
	if err != nil {
		t.Fatalf("AssignVariant failed: %v", err)
	}
	_ = assign3
}

func TestRecordEvent(t *testing.T) {
	svc := NewABTestService()

	exp, _ := svc.CreateExperiment(context.Background(), "test_exp2", []model.ABVariant{
		{Name: "a", Weight: 0.5},
		{Name: "b", Weight: 0.5},
	})

	svc.AssignVariant(context.Background(), exp.ID, "user_1")
	svc.AssignVariant(context.Background(), exp.ID, "user_2")

	err := svc.RecordEvent(context.Background(), exp.ID, "user_1", "a", "conversion")
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	err = svc.RecordEvent(context.Background(), exp.ID, "user_2", "a", "impression")
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}
}

func TestUniformDistribution(t *testing.T) {
	svc := NewABTestService()

	exp, _ := svc.CreateExperiment(context.Background(), "dist_test", []model.ABVariant{
		{Name: "a", Weight: 0.5},
		{Name: "b", Weight: 0.5},
	})

	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		userID := fmt.Sprintf("user_%d", i)
		assign, err := svc.AssignVariant(context.Background(), exp.ID, userID)
		if err != nil {
			t.Fatalf("AssignVariant failed: %v", err)
		}
		counts[assign.Variant]++
	}

	for variant, count := range counts {
		ratio := float64(count) / 1000.0
		if ratio < 0.35 || ratio > 0.65 {
			t.Fatalf("variant %s ratio %f is outside expected range", variant, ratio)
		}
	}
}

func TestGetResults(t *testing.T) {
	svc := NewABTestService()

	exp, _ := svc.CreateExperiment(context.Background(), "results_test", []model.ABVariant{
		{Name: "control", Weight: 0.5},
		{Name: "treatment", Weight: 0.5},
	})

	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user_%d", i)
		assign, _ := svc.AssignVariant(context.Background(), exp.ID, userID)
		if i < 30 {
			svc.RecordEvent(context.Background(), exp.ID, userID, assign.Variant, "conversion")
		}
	}

	results, err := svc.GetResults(context.Background(), exp.ID)
	if err != nil {
		t.Fatalf("GetResults failed: %v", err)
	}
	if len(results.Variants) != 2 {
		t.Fatalf("expected 2 variant results, got %d", len(results.Variants))
	}
	for _, vr := range results.Variants {
		if vr.Count == 0 {
			t.Fatalf("expected non-zero count for variant %s", vr.Variant)
		}
	}
}
