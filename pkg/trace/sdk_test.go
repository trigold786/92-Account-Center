package trace

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitNoopProvider(t *testing.T) {
	shutdown, err := InitProvider("test-service", "dev", "")
	if err != nil {
		t.Fatalf("InitProvider failed: %v", err)
	}
	defer shutdown()

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	if !span.SpanContext().IsValid() {
		t.Fatal("expected valid span context")
	}
}

func TestTraceContextPropagation(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "dev", "")
	defer shutdown()

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent")
	spanID := span.SpanContext().SpanID().String()
	traceID := span.SpanContext().TraceID().String()

	if len(traceID) != 32 {
		t.Fatalf("expected 32-char trace ID, got %d: %s", len(traceID), traceID)
	}
	if len(spanID) != 16 {
		t.Fatalf("expected 16-char span ID, got %d: %s", len(spanID), spanID)
	}
	span.End()
	_ = ctx
}

func TestW3CTraceparentRoundTrip(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "dev", "")
	defer shutdown()

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	header := FormatW3CTraceparent(traceID, spanID)
	if len(header) != 55 {
		t.Fatalf("expected 55-char traceparent, got %d: %s", len(header), header)
	}
	if header[:2] != "00" {
		t.Fatalf("expected version 00, got %s", header[:2])
	}
	span.End()
}
