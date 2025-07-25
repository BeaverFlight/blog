FROM golang:1.24 AS builder

WORKDIR /app


COPY . .

RUN go build -o blog .

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/blog /app/blog
COPY --from=builder /app/pkg/dbwork/migrations /app/pkg/dbwork/migrations
EXPOSE 8080

CMD ["/app/blog"]
