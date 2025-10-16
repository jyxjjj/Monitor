#!/bin/bash
# Integration test script for Monitor

set -e

echo "Monitor Integration Test"
echo "========================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8765"
PASSWORD="test123"
TEST_DIR="/tmp/monitor-test-$$"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    [ -n "$SERVER_PID" ] && kill $SERVER_PID 2>/dev/null || true
    [ -n "$AGENT_PID" ] && kill $AGENT_PID 2>/dev/null || true
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Create test directory
mkdir -p "$TEST_DIR"

# Check if binaries exist
if [ ! -f "./monitor-server" ] || [ ! -f "./monitor-agent" ]; then
    echo -e "${RED}Error: Binaries not found. Run 'make build' first.${NC}"
    exit 1
fi

echo -e "${YELLOW}Step 1: Creating test configuration...${NC}"

# Create server config
cat > "$TEST_DIR/server-config.json" << EOF
{
  "server_addr": ":8765",
  "tls_cert_file": "",
  "tls_key_file": "",
  "db_path": "$TEST_DIR/monitor.db",
  "admin_password": "$PASSWORD",
  "smtp_host": "",
  "smtp_port": 587,
  "smtp_user": "",
  "smtp_password": "",
  "email_from": "",
  "alert_email": ""
}
EOF

# Create agent config
cat > "$TEST_DIR/agent-config.json" << EOF
{
  "server_url": "$SERVER_URL",
  "agent_id": "test-agent",
  "agent_name": "Test Agent",
  "report_interval": 2,
  "tls_skip_verify": true
}
EOF

echo -e "${GREEN}✓ Configuration created${NC}"

echo -e "\n${YELLOW}Step 2: Starting server...${NC}"
./monitor-server --config "$TEST_DIR/server-config.json" > "$TEST_DIR/server.log" 2>&1 &
SERVER_PID=$!
sleep 3

if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${RED}✗ Server failed to start${NC}"
    cat "$TEST_DIR/server.log"
    exit 1
fi
echo -e "${GREEN}✓ Server started (PID: $SERVER_PID)${NC}"

echo -e "\n${YELLOW}Step 3: Testing login...${NC}"
TOKEN=$(curl -s -X POST "$SERVER_URL/api/login" \
    -H "Content-Type: application/json" \
    -d "{\"password\":\"$PASSWORD\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Login successful${NC}"

echo -e "\n${YELLOW}Step 4: Starting agent...${NC}"
./monitor-agent --config "$TEST_DIR/agent-config.json" > "$TEST_DIR/agent.log" 2>&1 &
AGENT_PID=$!
sleep 5

if ! kill -0 $AGENT_PID 2>/dev/null; then
    echo -e "${RED}✗ Agent failed to start${NC}"
    cat "$TEST_DIR/agent.log"
    exit 1
fi
echo -e "${GREEN}✓ Agent started (PID: $AGENT_PID)${NC}"

echo -e "\n${YELLOW}Step 5: Checking agent registration...${NC}"
AGENTS=$(curl -s "$SERVER_URL/api/agents" -H "Authorization: Bearer $TOKEN")
if echo "$AGENTS" | grep -q "test-agent"; then
    echo -e "${GREEN}✓ Agent registered successfully${NC}"
else
    echo -e "${RED}✗ Agent not found${NC}"
    echo "Response: $AGENTS"
    exit 1
fi

echo -e "\n${YELLOW}Step 6: Checking metrics data...${NC}"
sleep 5
METRICS=$(curl -s "$SERVER_URL/api/metrics/test-agent" -H "Authorization: Bearer $TOKEN")
if echo "$METRICS" | grep -q "cpu_percent"; then
    echo -e "${GREEN}✓ Metrics data found${NC}"
else
    echo -e "${RED}✗ No metrics data${NC}"
    echo "Response: $METRICS"
    exit 1
fi

echo -e "\n${YELLOW}Step 7: Creating alert rule...${NC}"
RULE=$(curl -s -X POST "$SERVER_URL/api/alert-rules" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "agent_id": "test-agent",
        "metric_type": "cpu",
        "threshold": 80,
        "operator": "gt",
        "duration": 5,
        "enabled": true,
        "description": "Test CPU alert"
    }')

if echo "$RULE" | grep -q '"id"'; then
    echo -e "${GREEN}✓ Alert rule created${NC}"
else
    echo -e "${RED}✗ Failed to create alert rule${NC}"
    echo "Response: $RULE"
    exit 1
fi

echo -e "\n${YELLOW}Step 8: Verifying alert rules...${NC}"
RULES=$(curl -s "$SERVER_URL/api/alert-rules" -H "Authorization: Bearer $TOKEN")
if echo "$RULES" | grep -q "Test CPU alert"; then
    echo -e "${GREEN}✓ Alert rule verified${NC}"
else
    echo -e "${RED}✗ Alert rule not found${NC}"
    exit 1
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}All tests passed successfully! ✓${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Test artifacts saved in: $TEST_DIR"
echo "Server log: $TEST_DIR/server.log"
echo "Agent log: $TEST_DIR/agent.log"
echo ""
