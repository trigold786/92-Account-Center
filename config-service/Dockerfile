FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY config-service/go.mod config-service/go.sum ./config-service/
COPY pkg/config/go.mod ./pkg/config/
RUN cd config-service && go mod download

COPY config-service/ ./config-service/
COPY pkg/config/ ./pkg/config/
RUN cd config-service && CGO_ENABLED=0 GOOS=linux go build -o /config-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add wget ca-certificates
COPY --from=builder /config-service /config-service

USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider http://localhost:30315/health || exit 1

EXPOSE 30315

CMD ["/config-service"]
