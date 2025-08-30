#!/bin/bash

echo "================================"
echo "Direct Sandbox Test"
echo "================================"
echo ""

# Build first
make build > /dev/null 2>&1

# Test 1: Create a test directory and try to write outside it
echo "Test 1: Testing file write restrictions"
echo "----------------------------------------"

# Create test directory
TEST_DIR="/tmp/rigel_sandbox_test_$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"
echo "Working directory: $(pwd)"

# Create a simple Go program that tries to write files
cat > test_write.go << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    // Try to write to /tmp (outside current directory)
    err := os.WriteFile("/tmp/test_forbidden.txt", []byte("This should fail"), 0644)
    if err != nil {
        fmt.Printf("✓ Cannot write to /tmp: %v\n", err)
    } else {
        fmt.Println("✗ Successfully wrote to /tmp (sandbox not working!)")
        os.Remove("/tmp/test_forbidden.txt")
    }

    // Try to write to current directory
    err = os.WriteFile("./test_allowed.txt", []byte("This should work"), 0644)
    if err != nil {
        fmt.Printf("✗ Cannot write to current directory: %v\n", err)
    } else {
        fmt.Println("✓ Successfully wrote to current directory")
    }
}
EOF

# Compile the test program
go build -o test_write test_write.go

echo ""
echo "Running WITHOUT sandbox:"
./test_write

echo ""
echo "Running WITH sandbox:"
# Create a simple sandbox profile
cat > sandbox.sb << EOF
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

sandbox-exec -f sandbox.sb ./test_write

echo ""
echo "Test 2: Testing Rigel with sandbox"
echo "------------------------------------"

# Test Rigel's sandbox mode
echo "Testing if Rigel sandbox is active..."
/Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel << 'EOF' 2>&1 | head -5
test
EOF

echo ""
echo "Test 3: Interactive verification"
echo "---------------------------------"
echo "To manually verify sandbox is working:"
echo "1. Run: ./bin/rigel"
echo "2. In another terminal, find the process: ps aux | grep rigel"
echo "3. Check sandbox status: sandbox-exec -p '(version 1)(allow default)' ps aux | grep rigel"
echo ""

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo "Test complete!"
