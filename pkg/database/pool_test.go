package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestPoolMetrics(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=test dbname=test sslmode=disable")
	if err != nil {
		t.Skipf("skipping: cannot open db: %v", err)
	}
	defer db.Close()

	cfg := PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	pool := NewOptimizedPool(db, cfg)
	metrics := pool.GetMetrics()

	if metrics.OpenConnections < 0 {
		t.Fatal("expected non-negative open connections")
	}
	if metrics.InUse < 0 {
		t.Fatal("expected non-negative in-use connections")
	}
}

func TestPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()
	if cfg.MaxOpenConns != 25 {
		t.Fatalf("expected 25 max open, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Fatalf("expected 10 max idle, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("expected 30m lifetime, got %v", cfg.ConnMaxLifetime)
	}
}

func TestPromMetricsText(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=test dbname=test sslmode=disable")
	if err != nil {
		t.Skipf("skipping: cannot open db: %v", err)
	}
	defer db.Close()

	pool := NewOptimizedPool(db, DefaultPoolConfig())
	text := pool.PromMetricsText()
	if text == "" {
		t.Fatal("expected non-empty metrics text")
	}
}
