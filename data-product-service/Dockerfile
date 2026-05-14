FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 30314
USER nobody
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30314/health || exit 1
ENTRYPOINT ["./main"]
