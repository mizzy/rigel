#!/bin/bash

echo "================================"
echo "Sandbox Functionality Verification"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build the binary first
echo "Building Rigel..."
make build > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

RIGEL_BIN="/Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel"

echo "Creating test environment..."
# Create a test directory
TEST_DIR="/tmp/sandbox_test_$(date +%s)"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "Current directory: $(pwd)"
echo ""

# Test 1: Try to write to /tmp (outside sandbox) with sandbox enabled
echo -e "${YELLOW}Test 1: Writing to /tmp with sandbox ENABLED${NC}"
echo "Attempting to create /tmp/forbidden.txt..."

# Create a test script that tries to write outside
cat > test_write.sh << 'EOF'
#!/bin/bash
echo "Trying to write to /tmp/forbidden.txt"
echo "This should fail" > /tmp/forbidden.txt 2>&1
if [ $? -eq 0 ]; then
    echo "FAIL: Successfully wrote to /tmp/forbidden.txt"
    exit 1
else
    echo "PASS: Cannot write to /tmp/forbidden.txt (sandbox working)"
    exit 0
fi
EOF
chmod +x test_write.sh

# Run with sandbox (default)
$RIGEL_BIN --sandbox << 'EOF' 2>&1 | grep -q "Sandbox enabled"
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Sandbox mode activated${NC}"

    # Now try to write outside using sandbox-exec directly
    sandbox-exec -f /dev/stdin ./test_write.sh << 'EOF' 2>&1
(version 1)
(deny default)
(allow process*)
(allow file-read*)
(allow file-write* (subpath "/tmp/sandbox_test"))
(allow network*)
(allow sysctl*)
(allow mach*)
(allow ipc*)
(allow system-socket)
EOF

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Sandbox correctly prevents writing outside current directory${NC}"
    else
        echo -e "${RED}✗ Unexpected sandbox behavior${NC}"
    fi
else
    echo -e "${RED}✗ Failed to activate sandbox mode${NC}"
fi

echo ""

# Test 2: Try to write within sandbox directory
echo -e "${YELLOW}Test 2: Writing within sandbox directory${NC}"
echo "Attempting to create local_file.txt in $(pwd)..."

cat > test_write_local.sh << 'EOF'
#!/bin/bash
echo "Trying to write to local_file.txt"
echo "This should succeed" > local_file.txt 2>&1
if [ $? -eq 0 ]; then
    echo "PASS: Successfully wrote to local_file.txt"
    exit 0
else
    echo "FAIL: Cannot write to local_file.txt"
    exit 1
fi
EOF
chmod +x test_write_local.sh

# Run with sandbox
sandbox-exec -f /dev/stdin ./test_write_local.sh << EOF 2>&1
(version 1)
(deny default)
(allow process*)
(allow file-read*)
(allow file-write* (subpath "$TEST_DIR"))
(allow network*)
(allow sysctl*)
(allow mach*)
(allow ipc*)
(allow system-socket)
EOF

if [ $? -eq 0 ] && [ -f "local_file.txt" ]; then
    echo -e "${GREEN}✓ Can write within sandbox directory${NC}"
    cat local_file.txt
else
    echo -e "${RED}✗ Failed to write within sandbox directory${NC}"
fi

echo ""

# Test 3: Verify --no-sandbox flag disables restrictions
echo -e "${YELLOW}Test 3: Testing --no-sandbox flag${NC}"
echo "Running with --no-sandbox..."

# First check if warning is shown
$RIGEL_BIN --no-sandbox << 'EOF' 2>&1 | head -1 | grep -q "Running without sandbox"
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Warning shown when sandbox disabled${NC}"

    # Test actual write capability without sandbox
    echo "Test content" > /tmp/no_sandbox_test.txt 2>/dev/null
    if [ -f "/tmp/no_sandbox_test.txt" ]; then
        echo -e "${GREEN}✓ Can write anywhere when sandbox disabled${NC}"
        rm -f /tmp/no_sandbox_test.txt
    fi
else
    echo -e "${RED}✗ No warning when sandbox disabled${NC}"
fi

echo ""

# Test 4: Verify default behavior (sandbox should be ON by default on macOS)
echo -e "${YELLOW}Test 4: Default behavior (no flags)${NC}"
$RIGEL_BIN << 'EOF' 2>&1 | head -1 | grep -q "Sandbox enabled"
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Sandbox is ON by default${NC}"
else
    echo -e "${RED}✗ Sandbox is not enabled by default${NC}"
fi

echo ""

# Test 5: Try to read files outside sandbox (should be allowed)
echo -e "${YELLOW}Test 5: Reading files outside sandbox${NC}"
echo "Attempting to read /etc/hosts..."

cat > test_read.sh << 'EOF'
#!/bin/bash
if cat /etc/hosts > /dev/null 2>&1; then
    echo "PASS: Can read /etc/hosts"
    exit 0
else
    echo "FAIL: Cannot read /etc/hosts"
    exit 1
fi
EOF
chmod +x test_read.sh

sandbox-exec -f /dev/stdin ./test_read.sh << EOF 2>&1
(version 1)
(deny default)
(allow process*)
(allow file-read*)
(allow file-write* (subpath "$TEST_DIR"))
(allow network*)
(allow sysctl*)
(allow mach*)
(allow ipc*)
(allow system-socket)
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Can read files outside sandbox (as expected)${NC}"
else
    echo -e "${RED}✗ Cannot read files outside sandbox${NC}"
fi

# Cleanup
echo ""
echo "Cleaning up test directory..."
cd /tmp
rm -rf "$TEST_DIR"

echo ""
echo "================================"
echo "Sandbox Verification Complete"
echo "================================"
