#!/bin/bash

echo "=== Testing real terminal cumulative indent issue ==="
echo "Building test binary..."
RIGEL_TEST_MODE=1 PROVIDER=ollama go build -o rigel-test-env cmd/rigel/main.go

if [ ! -f "rigel-test-env" ]; then
    echo "Failed to build test binary"
    exit 1
fi

echo "Clearing debug log..."
rm -f /tmp/rigel_debug.log

echo ""
echo "=== MANUAL TEST INSTRUCTIONS ==="
echo "1. Run: RIGEL_TEST_MODE=1 PROVIDER=ollama ./rigel-test-env --termflow"
echo "2. Type: aaaa"
echo "3. Press Ctrl+J"
echo "4. Type: bbbb" 
echo "5. Press Ctrl+J"
echo "6. Type: cccc"
echo "7. Press Ctrl+J"
echo "8. Type: dddd"
echo "9. Press Ctrl+C twice to exit"
echo ""
echo "After test, check debug log with: cat /tmp/rigel_debug.log"
echo "Look for cumulative indentation in the terminal display"
echo ""

# Create a script to run the test in a subshell
cat > run_test.sh << 'EOF'
#!/bin/bash
export RIGEL_TEST_MODE=1
export PROVIDER=ollama
echo "Starting rigel with debug logging..."
./rigel-test-env --termflow
echo "Test completed. Debug log:"
if [ -f "/tmp/rigel_debug.log" ]; then
    cat /tmp/rigel_debug.log
else
    echo "No debug log found"
fi
EOF

chmod +x run_test.sh
echo "Test script created: run_test.sh"
echo "Run: ./run_test.sh and follow the manual test instructions above"