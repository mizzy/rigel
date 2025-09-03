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
