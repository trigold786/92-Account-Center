package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
)

func TestCorrelationInContext(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	corrHandler := NewCorrelationHandler(handler)
	logger := slog.New(corrHandler)

	ctx := WithTraceID(context.Background(), "trace-123")
	ctx = WithSpanID(ctx, "span-456")

	logger.InfoContext(ctx, "test message")

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte(`"trace_id":"trace-123"`)) {
		t.Fatalf("expected trace_id in output, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"span_id":"span-456"`)) {
		t.Fatalf("expected span_id in output, got: %s", output)
	}
}

func TestCorrelationEmptyContext(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	corrHandler := NewCorrelationHandler(handler)
	logger := slog.New(corrHandler)

	logger.InfoContext(context.Background(), "no correlation")

	output := buf.String()
	if bytes.Contains(buf.Bytes(), []byte(`"trace_id"`)) {
		t.Fatalf("expected no trace_id in output, got: %s", output)
	}
}

func TestWithTraceSpanID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "t1")
	ctx = WithSpanID(ctx, "s1")

	if getTraceID(ctx) != "t1" {
		t.Fatal("expected t1")
	}
	if getSpanID(ctx) != "s1" {
		t.Fatal("expected s1")
	}
}
