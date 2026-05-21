package discovery

import (
	"context"
	"testing"
)

func TestRegister(t *testing.T) {
	reg := NewInMemoryRegistry()

	err := reg.Register(context.Background(), &ServiceInstance{
		ID:      "svc-1",
		Name:    "auth-service",
		Address: "localhost",
		Port:    30302,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	instances, err := reg.Discover(context.Background(), "auth-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "svc-1" {
		t.Fatalf("expected svc-1, got %s", instances[0].ID)
	}
}

func TestDeregister(t *testing.T) {
	reg := NewInMemoryRegistry()

	reg.Register(context.Background(), &ServiceInstance{ID: "svc-1", Name: "test", Address: "localhost", Port: 8080})
	reg.Register(context.Background(), &ServiceInstance{ID: "svc-2", Name: "test", Address: "localhost", Port: 8081})

	err := reg.Deregister(context.Background(), "svc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	instances, _ := reg.Discover(context.Background(), "test")
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance after deregister, got %d", len(instances))
	}
	if instances[0].ID != "svc-2" {
		t.Fatalf("expected svc-2, got %s", instances[0].ID)
	}
}

func TestDiscover(t *testing.T) {
	reg := NewInMemoryRegistry()

	reg.Register(context.Background(), &ServiceInstance{ID: "a-1", Name: "svc-a", Address: "localhost", Port: 1})
	reg.Register(context.Background(), &ServiceInstance{ID: "a-2", Name: "svc-a", Address: "localhost", Port: 2})
	reg.Register(context.Background(), &ServiceInstance{ID: "b-1", Name: "svc-b", Address: "localhost", Port: 3})

	instances, _ := reg.Discover(context.Background(), "svc-a")
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances for svc-a, got %d", len(instances))
	}

	instances, _ = reg.Discover(context.Background(), "svc-b")
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance for svc-b, got %d", len(instances))
	}

	instances, _ = reg.Discover(context.Background(), "nonexistent")
	if instances != nil {
		t.Fatal("expected nil for nonexistent service")
	}
}

func TestHealthCheck(t *testing.T) {
	reg := NewInMemoryRegistry()

	reg.Register(context.Background(), &ServiceInstance{ID: "svc-1", Name: "test", Address: "localhost", Port: 8080})

	err := reg.HealthCheck(context.Background(), "svc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = reg.HealthCheck(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent service")
	}
}

func TestRegisterValidation(t *testing.T) {
	reg := NewInMemoryRegistry()

	err := reg.Register(context.Background(), &ServiceInstance{ID: "", Name: "test"})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}

	err = reg.Register(context.Background(), &ServiceInstance{ID: "test", Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}
