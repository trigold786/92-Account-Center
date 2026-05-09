FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /risk-detection-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /risk-detection-service /usr/local/bin/risk-detection-service

EXPOSE 8086

CMD ["risk-detection-service"]