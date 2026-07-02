package saga

import "context"

type StepAction func(ctx context.Context) error

// SagaStep is a named unit of work with an execute and optional compensate action.
type SagaStep struct {
	Name       string
	Execute    StepAction
	Compensate StepAction
	executed   bool
}

func NewStep(name string, execute, compensate StepAction) *SagaStep {
	return &SagaStep{Name: name, Execute: execute, Compensate: compensate}
}
