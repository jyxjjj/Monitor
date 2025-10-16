.PHONY: all build server agent frontend clean install-deps test

all: build

# Install Go dependencies
install-deps:
	go mod download

# Build server
server:
	go build -o monitor-server ./cmd/server

# Build agent
agent:
	go build -o monitor-agent ./cmd/agent

# Build frontend
frontend:
	cd frontend && npm install && npm run build

# Build all
build: server agent

# Build all including frontend
build-all: server agent frontend

# Run server
run-server:
	./monitor-server

# Run agent
run-agent:
	./monitor-agent

# Test
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f monitor-server monitor-agent
	rm -rf frontend/build
	rm -rf frontend/node_modules
	rm -f monitor.db*

# Generate TLS certificates for development
gen-cert:
	openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"

# Cross-compile for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o monitor-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=amd64 go build -o monitor-agent-linux-amd64 ./cmd/agent

build-windows:
	GOOS=windows GOARCH=amd64 go build -o monitor-server-windows-amd64.exe ./cmd/server
	GOOS=windows GOARCH=amd64 go build -o monitor-agent-windows-amd64.exe ./cmd/agent

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o monitor-server-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o monitor-agent-darwin-amd64 ./cmd/agent
	GOOS=darwin GOARCH=arm64 go build -o monitor-server-darwin-arm64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o monitor-agent-darwin-arm64 ./cmd/agent

build-all-platforms: build-linux build-windows build-macos

# Help
help:
	@echo "Available targets:"
	@echo "  all              - Build server and agent (default)"
	@echo "  install-deps     - Install Go dependencies"
	@echo "  server           - Build server only"
	@echo "  agent            - Build agent only"
	@echo "  frontend         - Build frontend only"
	@echo "  build            - Build server and agent"
	@echo "  build-all        - Build server, agent, and frontend"
	@echo "  run-server       - Run server"
	@echo "  run-agent        - Run agent"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  gen-cert         - Generate self-signed TLS certificate"
	@echo "  build-all-platforms - Cross-compile for all platforms"
