FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /session-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /session-service /usr/local/bin/session-service

EXPOSE 30307
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30307/health || exit 1

CMD ["session-service"]
