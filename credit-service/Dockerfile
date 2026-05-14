FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /credit-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /credit-service /usr/local/bin/credit-service

EXPOSE 30312
USER nobody

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --spider -q http://127.0.0.1:30312/health || exit 1

CMD ["credit-service"]
