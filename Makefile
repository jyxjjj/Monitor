.PHONY: all build server agent frontend clean deps test

all: server agent frontend

all-platforms: build-linux build-windows build-macos frontend

# Install Go dependencies
deps:
	go mod tidy
	cd frontend && npm install && cd ..

# Build server
server:
	CGO_ENABLED=1 go build -o monitor-server ./cmd/server

# Build agent
agent:
	CGO_ENABLED=1 go build -o monitor-agent ./cmd/agent

# Build frontend
frontend:
	cd frontend && npm run build && cd ..

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
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o monitor-server-linux-amd64 ./cmd/server
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o monitor-agent-linux-amd64 ./cmd/agent

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o monitor-server-windows-amd64.exe ./cmd/server
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o monitor-agent-windows-amd64.exe ./cmd/agent

build-macos:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o monitor-server-darwin-arm64 ./cmd/server
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o monitor-agent-darwin-arm64 ./cmd/agent
