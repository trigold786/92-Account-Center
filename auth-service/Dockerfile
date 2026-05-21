FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /auth-service /usr/local/bin/auth-service

EXPOSE 30302
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30302/health || exit 1

CMD ["auth-service"]
