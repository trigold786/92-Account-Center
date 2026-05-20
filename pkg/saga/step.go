package saga

import "context"

type StepAction func(ctx context.Context) error

type SagaStep struct {
	Name       string
	Execute    StepAction
	Compensate StepAction
	executed   bool
}

func NewStep(name string, execute, compensate StepAction) *SagaStep {
	return &SagaStep{Name: name, Execute: execute, Compensate: compensate}
}
