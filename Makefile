.PHONY: build test install

build:
	go build -o byte-cli ./cmd/byte-cli

test:
	go test ./...

install:
	go install ./cmd/byte-cli
