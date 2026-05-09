FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /device-fingerprint-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /device-fingerprint-service /usr/local/bin/device-fingerprint-service

EXPOSE 8089

CMD ["device-fingerprint-service"]