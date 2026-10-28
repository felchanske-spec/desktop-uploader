ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: run test

run:
	@echo "Launching desktop-uploader with environment variables..."
	go run cmd/uploader-backend/main.go

test:
	@echo "Running all unit tests safely isolated..."
	go test -v ./...

build-service:
	@echo "Compiling native macOS system service binary"
	go build -o bin/uploader-service cmd/uploader-service/main.go