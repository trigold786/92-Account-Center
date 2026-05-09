FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /audit-log-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /audit-log-service /usr/local/bin/audit-log-service

EXPOSE 8085

CMD ["audit-log-service"]