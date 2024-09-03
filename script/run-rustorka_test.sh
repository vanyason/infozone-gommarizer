#!/usr/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to that directory
cd "$SCRIPT_DIR" || exit

# Go to the parent directory - project root
cd ..

# Recreate a temporary directory for test files
TEST_DIR="tmp/rustorka_test"
rm -rf "$TEST_DIR" && mkdir -p "$TEST_DIR"

# Go to the test directory
cd "$TEST_DIR" || exit

# Run the tests
go run ../../test/rustorka_test/main.go && sed -i '/kot.png/d' rustorka-*html