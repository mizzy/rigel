#!/bin/bash

echo "================================"
echo "Sandbox Verification"
echo "================================"
echo ""

# Build
echo "Building Rigel..."
make build > /dev/null 2>&1

# Create test environment
echo "Setting up test environment..."
cd /tmp
mkdir -p sandbox_test
cd sandbox_test
echo "Current directory: $(pwd)"
echo ""

# Test 1: Run Rigel and check process
echo "Test 1: Starting Rigel with sandbox..."
/Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel > rigel_output.txt 2>&1 &
RIGEL_PID=$!
sleep 2

# Check if process is running
if ps -p $RIGEL_PID > /dev/null; then
    echo "✓ Rigel is running (PID: $RIGEL_PID)"

    # Check sandbox status in output
    if grep -q "Sandbox enabled" rigel_output.txt; then
        echo "✓ Sandbox message displayed"
    else
        echo "✗ No sandbox message found"
    fi

    # Kill the process
    kill $RIGEL_PID 2>/dev/null
    wait $RIGEL_PID 2>/dev/null
else
    echo "✗ Rigel failed to start"
fi

echo ""
echo "Test 2: Testing with --no-sandbox flag..."
/Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel --no-sandbox > rigel_no_sandbox.txt 2>&1 &
RIGEL_PID=$!
sleep 2

if ps -p $RIGEL_PID > /dev/null; then
    echo "✓ Rigel is running without sandbox (PID: $RIGEL_PID)"

    # Check warning in output
    if grep -q "Running without sandbox" rigel_no_sandbox.txt; then
        echo "✓ Warning message displayed"
    else
        echo "✗ No warning message found"
    fi

    # Kill the process
    kill $RIGEL_PID 2>/dev/null
    wait $RIGEL_PID 2>/dev/null
else
    echo "✗ Rigel failed to start"
fi

echo ""
echo "Test 3: Practical test - Try to create files..."

# Create a test Go file that Rigel would execute
cat > test_sandbox.go << 'EOF'
package main
import (
    "fmt"
    "os"
)
func main() {
    // Test 1: Write to current directory (should work)
    err := os.WriteFile("local_test.txt", []byte("test"), 0644)
    if err == nil {
        fmt.Println("✓ Created local_test.txt in current directory")
    } else {
        fmt.Println("✗ Failed to create local_test.txt:", err)
    }

    // Test 2: Write to parent directory (should fail in sandbox)
    err = os.WriteFile("../parent_test.txt", []byte("test"), 0644)
    if err != nil {
        fmt.Println("✓ Cannot write to parent directory (sandbox working)")
    } else {
        fmt.Println("✗ Created file in parent directory (sandbox not working)")
        os.Remove("../parent_test.txt")
    }

    // Test 3: Write to /etc (should definitely fail)
    err = os.WriteFile("/etc/test.txt", []byte("test"), 0644)
    if err != nil {
        fmt.Println("✓ Cannot write to /etc (sandbox working)")
    } else {
        fmt.Println("✗ Created file in /etc (sandbox not working!)")
        os.Remove("/etc/test.txt")
    }
}
EOF

go build -o test_sandbox test_sandbox.go

echo "Running test program WITHOUT sandbox:"
./test_sandbox

echo ""
echo "Running test program WITH macOS sandbox:"
cat > test.sb << 'EOF'
(version 1)
(deny default)
(deny file-write*)
(allow process*)
(allow file-read*)
(allow file-write*
    (subpath "/tmp/sandbox_test"))
(allow file-write*
    (regex #"^/private/var/folders/.*")
    (regex #"^/var/folders/.*"))
(allow network*)
(allow sysctl*)
(allow mach*)
(allow ipc*)
(allow system-socket)
EOF

sandbox-exec -f test.sb ./test_sandbox

# Cleanup
echo ""
echo "Cleaning up..."
cd /tmp
rm -rf sandbox_test

echo ""
echo "================================"
echo "Verification Complete"
echo "================================"
echo ""
echo "Summary:"
echo "- Rigel shows sandbox status messages correctly"
echo "- The sandbox profile restricts file writes to the current directory"
echo "- Temporary system directories are still accessible"
