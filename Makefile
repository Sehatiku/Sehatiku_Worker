.PHONY: run build setup-wa test tidy

run:
	go run ./cmd/worker

build:
	go build -o bin/worker ./cmd/worker

setup-wa:
	go run ./cmd/wa-setup

test:
	go test ./... -v

tidy:
	go mod tidy
