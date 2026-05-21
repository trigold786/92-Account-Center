package trace

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		tracer := otel.Tracer(serviceName)
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		if sc := span.SpanContext(); sc.HasTraceID() {
			tp := FormatW3CTraceparent(
				sc.TraceID().String(),
				sc.SpanID().String(),
			)
			c.Header("traceparent", tp)
		}

		c.Next()
	}
}
