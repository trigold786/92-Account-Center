FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /api-gateway /usr/local/bin/api-gateway

EXPOSE 30300
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30300/health || exit 1

CMD ["api-gateway"]
