FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /account-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /account-service /usr/local/bin/account-service

EXPOSE 30301
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30301/health || exit 1

CMD ["account-service"]
