APP_NAME := michael-connelly-api
BUILD_DIR := ./cmd

build:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(BUILD_DIR)

run:
	go run ./cmd/main.go

clean:
	rm -f $(APP_NAME)