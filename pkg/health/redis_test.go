package health

import (
	"context"
	"testing"
)

func TestRedisCheckerMissingRedis(t *testing.T) {
	rc := &RedisChecker{RedisNop: true}
	result := rc.Check(context.Background())
	if result.Status != StatusDown {
		t.Fatalf("expected down without Redis, got %v", result.Status)
	}
}

func TestRedisCheckerName(t *testing.T) {
	rc := &RedisChecker{}
	if rc.Name() != "redis" {
		t.Fatalf("unexpected name: %s", rc.Name())
	}
}
