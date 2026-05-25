# TaskFlow Makefile

APP_NAME    := taskflow
CMD_PATH    := ./cmd/app
MIGRATIONS  := ./migrations
DB_URL      ?= postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable

.PHONY: help build run test test-cover lint tidy \
        docker-up docker-down docker-build \
        migrate-up migrate-down migrate-create \
        proto swagger

help:
	@echo "Available targets:"
	@echo "  build           - Build the application binary"
	@echo "  run             - Run the application locally"
	@echo "  test            - Run unit tests"
	@echo "  test-cover      - Run tests with coverage"
	@echo "  tidy            - Tidy go modules"
	@echo "  docker-up       - Start docker-compose stack"
	@echo "  docker-down     - Stop docker-compose stack"
	@echo "  migrate-up      - Apply database migrations"
	@echo "  migrate-down    - Rollback last migration"
	@echo "  migrate-create  - Create new migration (NAME=...)"
	@echo "  proto           - Generate code from proto files"
	@echo "  swagger         - Generate swagger documentation"

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./... -race -count=1

test-cover:
	go test ./... -race -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

tidy:
	go mod tidy

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

docker-build:
	docker compose build

migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS) up

migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS) down 1

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(NAME)

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/notification.proto

swagger:
	swag init -g cmd/app/main.go -o docs
