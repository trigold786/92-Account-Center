package saga

import (
	"context"
	"errors"
	"testing"
)

func TestSagaStepExecute(t *testing.T) {
	executed := false
	step := NewStep("deduct_credits", func(ctx context.Context) error {
		executed = true
		return nil
	}, nil)
	err := step.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !executed {
		t.Fatal("step was not executed")
	}
}

func TestSagaStepCompensate(t *testing.T) {
	compensated := false
	step := NewStep("deduct_credits",
		func(ctx context.Context) error { return errors.New("execution failed") },
		func(ctx context.Context) error {
			compensated = true
			return nil
		},
	)
	err := step.Execute(context.Background())
	if err == nil {
		t.Fatal("expected execution error")
	}
	err = step.Compensate(context.Background())
	if err != nil {
		t.Fatalf("Compensate failed: %v", err)
	}
	if !compensated {
		t.Fatal("compensation was not executed")
	}
}
