ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: run test

run:
	@echo "Launching desktop-uploader with environment variables..."
	go run main.go

test:
	@echo "Running all unit tests safely isolated..."
	go test -v ./...
