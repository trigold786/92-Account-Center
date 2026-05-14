FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /data-product-service ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget
COPY --from=builder /data-product-service /usr/local/bin/data-product-service
EXPOSE 30314
USER nobody
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30314/health || exit 1
CMD ["data-product-service"]
