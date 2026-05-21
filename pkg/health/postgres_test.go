package health

import (
	"context"
	"testing"
)

func TestPostgresCheckerMissingDB(t *testing.T) {
	pc := &PostgresChecker{DBNop: true}
	result := pc.Check(context.Background())
	if result.Status != StatusDown {
		t.Fatalf("expected down without DB, got %v", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestPostgresCheckerName(t *testing.T) {
	pc := &PostgresChecker{}
	if pc.Name() != "postgres" {
		t.Fatalf("unexpected name: %s", pc.Name())
	}
}
