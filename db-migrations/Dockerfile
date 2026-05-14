FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.20.0

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY . /migrations
WORKDIR /migrations
ENTRYPOINT ["goose"]
