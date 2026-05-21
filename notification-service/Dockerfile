FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /notification-service ./cmd/main.go

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /notification-service /usr/local/bin/notification-service

EXPOSE 30311

USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30311/health || exit 1

CMD ["notification-service"]
