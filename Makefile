.PHONY: run build test docker-up docker-down clean

all: build

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/app ./cmd/server

test:
	go test ./...

# Docker commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

# Code quality
fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	docker-compose down -v
