APP_NAME := michael-connelly-api
BUILD_DIR := ./cmd

.PHONY: build run test clean

build:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(BUILD_DIR)

up:
	docker compose up -d

down:
	docker compose down

run: up
	go run ./cmd/main.go

test:
	go test ./internal/... -count=1

integration-tests:
	go test  ./test/... -count=1

clean:
	rm -f $(APP_NAME)