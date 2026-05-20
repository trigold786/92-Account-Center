package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type correlationCtxKey string

const (
	traceIDKey  correlationCtxKey = "trace_id"
	spanIDKey   correlationCtxKey = "span_id"
)

type CorrelationHandler struct {
	handler slog.Handler
}

func NewCorrelationHandler(h slog.Handler) *CorrelationHandler {
	return &CorrelationHandler{handler: h}
}

func (ch *CorrelationHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return ch.handler.Enabled(ctx, level)
}

func (ch *CorrelationHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID := getTraceID(ctx); traceID != "" {
		r.Add("trace_id", traceID)
	}
	if spanID := getSpanID(ctx); spanID != "" {
		r.Add("span_id", spanID)
	}
	return ch.handler.Handle(ctx, r)
}

func (ch *CorrelationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CorrelationHandler{handler: ch.handler.WithAttrs(attrs)}
}

func (ch *CorrelationHandler) WithGroup(name string) slog.Handler {
	return &CorrelationHandler{handler: ch.handler.WithGroup(name)}
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

func getTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

func getSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

func NewCorrelatedLogger(service string) *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	corrHandler := NewCorrelationHandler(baseHandler)
	return slog.New(corrHandler).With("service", service)
}

func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-Request-ID")
		}
		spanID := c.GetHeader("X-Span-ID")

		ctx := c.Request.Context()
		if traceID != "" {
			ctx = WithTraceID(ctx, traceID)
		}
		if spanID != "" {
			ctx = WithSpanID(ctx, spanID)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
