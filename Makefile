.PHONY: all frontend backend clean dev

# Default target
all: backend

# Build frontend
frontend:
	cd frontend && npm install && npm run build

# Build backend (includes frontend)
backend: frontend
	go build -o nft-ui .

# Development mode - run backend only (frontend uses vite dev server)
dev:
	go run .

# Clean build artifacts
clean:
	rm -rf frontend/dist frontend/node_modules nft-ui

# Install dependencies
deps:
	go mod tidy
	cd frontend && npm install

# Build for Linux (cross-compile)
build-linux:
	cd frontend && npm run build
	GOOS=linux GOARCH=amd64 go build -o nft-ui-linux-amd64 .

# Build for multiple platforms
build-all: frontend
	GOOS=linux GOARCH=amd64 go build -o nft-ui-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o nft-ui-linux-arm64 .
