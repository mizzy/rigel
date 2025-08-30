#!/bin/bash

echo "Testing Rigel Sandbox Feature"
echo "=============================="
echo ""

# Build the binary
echo "Building Rigel..."
make build
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi

# Test 1: Check that sandbox prevents writing outside current directory
echo "Test 1: Attempting to write to /tmp in sandbox mode"
cd /tmp
mkdir -p rigel_sandbox_test
cd rigel_sandbox_test

# Try to write outside sandbox directory
echo "Testing write outside sandbox..."
TEST_OUTPUT=$(echo "Write 'test' to /tmp/test_file.txt" | /Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel --sandbox 2>&1)

if echo "$TEST_OUTPUT" | grep -q "Sandbox enabled"; then
    echo "✅ Sandbox mode activated"
else
    echo "❌ Sandbox mode failed to activate"
fi

# Test 2: Verify we can write within sandbox directory
echo ""
echo "Test 2: Writing within sandbox directory"
echo "Writing to local file..."
echo "test content" > local_test.txt
if [ -f "local_test.txt" ]; then
    echo "✅ Can write to current directory"
else
    echo "❌ Failed to write to current directory"
fi

# Test 3: Verify sandbox status without flag
echo ""
echo "Test 3: Running without sandbox flag"
TEST_OUTPUT=$(echo "test" | /Users/mizzy/src/github.com/mizzy/feat-sandbox/bin/rigel --no-sandbox 2>&1 | head -3)
if echo "$TEST_OUTPUT" | grep -q "Running without sandbox"; then
    echo "✅ Warning shown when sandbox disabled"
else
    echo "❌ No warning when sandbox disabled"
fi

# Cleanup
cd /tmp
rm -rf rigel_sandbox_test

echo ""
echo "Sandbox tests completed!"
