package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewDefault(t *testing.T) {
	cb := New(3, time.Second)
	if cb.State() != StateClosed {
		t.Fatalf("expected closed, got %v", cb.State())
	}
	if cb.SuccessCount() != 0 || cb.FailureCount() != 0 {
		t.Fatalf("expected zero counts")
	}
}

func TestClosedToOpenOnFailures(t *testing.T) {
	cb := New(2, time.Minute)
	err1 := cb.Execute(func() error { return errors.New("fail") })
	if err1 == nil {
		t.Fatal("expected error")
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after 1 failure, got %v", cb.State())
	}
	err2 := cb.Execute(func() error { return errors.New("fail") })
	if err2 == nil {
		t.Fatal("expected error")
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected open after 2 failures, got %v", cb.State())
	}
}

func TestOpenRejectsRequests(t *testing.T) {
	cb := New(1, time.Minute)
	cb.Execute(func() error { return errors.New("fail") })
	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestOpenToHalfOpenAfterTimeout(t *testing.T) {
	cb := New(1, 50*time.Millisecond)
	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(80 * time.Millisecond)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil after timeout, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after success in half-open, got %v", cb.State())
	}
}

func TestHalfOpenRequiresMaxSuccesses(t *testing.T) {
	cb := NewWithOptions(Options{
		MaxFailures: 3,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 2,
	})
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("fail") })
	}
	time.Sleep(80 * time.Millisecond)

	err1 := cb.Execute(func() error { return nil })
	if err1 != nil {
		t.Fatalf("expected first half-open probe to pass, got %v", err1)
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected half-open after 1/2 successes, got %v", cb.State())
	}

	err2 := cb.Execute(func() error { return nil })
	if err2 != nil {
		t.Fatalf("expected second half-open probe to pass, got %v", err2)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after 2/2 successes, got %v", cb.State())
	}
}

func TestHalfOpenFailureReopens(t *testing.T) {
	cb := NewWithOptions(Options{
		MaxFailures: 3,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 2,
	})
	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("fail") })
	}
	time.Sleep(80 * time.Millisecond)

	cb.Execute(func() error { return nil })
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected half-open after 1/2 successes, got %v", cb.State())
	}

	err := cb.Execute(func() error { return errors.New("fail") })
	if err == nil {
		t.Fatal("expected failure in half-open probe")
	}
	if cb.State() != StateOpen {
		t.Fatalf("expected open after half-open failure, got %v", cb.State())
	}
}

func TestConcurrency(t *testing.T) {
	cb := New(5, time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}()
	}
	wg.Wait()
	if cb.State() != StateClosed {
		t.Fatalf("expected closed after all success, got %v", cb.State())
	}
}

func TestOnStateChangeCallback(t *testing.T) {
	var transitions []string
	cb := NewWithOptions(Options{
		MaxFailures: 1,
		Timeout:     50 * time.Millisecond,
		HalfOpenMax: 1,
		OnStateChange: func(from, to State) {
			transitions = append(transitions, from.String()+"->"+to.String())
		},
	})

	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(80 * time.Millisecond)
	cb.Execute(func() error { return nil })

	if len(transitions) < 2 {
		t.Fatalf("expected at least 2 transitions, got %v", transitions)
	}
}

func TestMetrics(t *testing.T) {
	cb := New(3, time.Minute)
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return errors.New("fail") })
	cb.Execute(func() error { return nil })

	if cb.SuccessCount() != 2 {
		t.Fatalf("expected 2 successes, got %d", cb.SuccessCount())
	}
	if cb.FailureCount() != 1 {
		t.Fatalf("expected 1 failure, got %d", cb.FailureCount())
	}
}
