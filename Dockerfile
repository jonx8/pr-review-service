# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/server


FROM scratch

COPY --from=builder /app/main /

USER 1000:1000

EXPOSE 8080

CMD ["/main"]
