# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/server

FROM alpine:3.22.2

WORKDIR /app

COPY --from=builder --chown=1000:1000 /app/main ./main

COPY --chown=1000:1000 migrations/ ./migrations/

USER 1000:1000

EXPOSE 8080

CMD ["sh", "-c", "./main"]

