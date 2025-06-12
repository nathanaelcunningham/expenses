.PHONY: proto build run-server run-client clean

# Generate protobuf files
proto:
	buf generate proto

# Build server and client
build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Run server
run-server:
	go run ./cmd/server

# Run client
run-client:
	go run ./cmd/client

# Clean generated files and binaries
clean:
	rm -rf gen/*.pb.go
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy

# Generate proto and build
all: proto build