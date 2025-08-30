#!/bin/bash

# Test script for rigel generate command

# Set up test environment
cp .env.example .env

echo "=== Testing Rigel Code Generation ==="
echo

# Test 1: Generate a simple function
echo "Test 1: Generate fibonacci function"
./rigel generate "Create a Go function that calculates the nth fibonacci number"
echo
echo "---"
echo

# Test 2: Generate with more complex requirements
echo "Test 2: Generate HTTP handler"
./rigel generate "Create a Go HTTP handler that accepts JSON input with name and age fields and returns a greeting"
echo
echo "---"
echo

# Test 3: Analyze existing code
echo "Test 3: Analyze main.go"
./rigel analyze cmd/rigel/main.go
echo

echo "=== Tests completed ==="
