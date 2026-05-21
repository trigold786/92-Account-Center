FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-service ./cmd/main.go

FROM alpine:3.23

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /payment-service /usr/local/bin/payment-service

EXPOSE 30316
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30316/health || exit 1

CMD ["payment-service"]
