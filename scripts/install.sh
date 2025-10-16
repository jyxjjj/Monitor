#!/bin/bash
# Monitor installation script

set -e

echo "Monitor Installation Script"
echo "==========================="
echo ""

# Check for required commands
command -v go >/dev/null 2>&1 || { echo "Error: Go is not installed. Please install Go 1.21+"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Warning: Node.js is not installed. Frontend will not be built."; }

# Build server and agent
echo "Building server and agent..."
make build

# Build frontend if Node.js is available
if command -v node >/dev/null 2>&1; then
    echo "Building frontend..."
    make frontend
fi

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Run './monitor-server' to start the server (will generate default config)"
echo "2. Edit server-config.json with your settings"
echo "3. Run './monitor-server' again to start with your config"
echo "4. Run './monitor-agent' on each server you want to monitor"
echo ""
