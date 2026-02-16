# BotCall Makefile
.PHONY: all build server pwa clean test run

all: build

build: server

# Build Go server
server:
	cd server && go build -o ../bin/botcall-server ./cmd/botcall-server

# Run development server
run:
	cd server && go run ./cmd/botcall-server

# Run tests
test:
	cd server && go test ./...

# Dependencies
deps:
	cd server && go mod tidy
	cd server && go mod download

# Clean build artifacts
clean:
	rm -rf bin/

# Docker
docker-build:
	docker build -t botcall-server:latest -f server/Dockerfile .

docker-run:
	docker run -p 8080:8080 botcall-server:latest

# Development setup
dev:
	@mkdir -p bin

# Full dev stack
dev-up:
	@echo "Starting BotCall server..."
	make run

# Lint
check:
	cd server && go vet ./...
	cd server && go fmt ./...
