#!/bin/bash
set -e

echo "=== Moistello Load Test ==="
echo ""

# Test 1: Sequential sequences
echo "--- Test 1: Sequential Sequences ---"
go test -v -run TestLoad_SequentialContributions ./tests/loadtest/ -timeout 120s -count=1 2>&1

echo ""
echo "--- Test 2: Concurrent Sequences ---"
go test -v -run TestLoad_ConcurrentSequences ./tests/loadtest/ -timeout 120s -count=1 2>&1

echo ""
echo "--- Test 3: Transaction Builder Stress ---"
go test -v -run TestLoad_TransactionBuilderStress ./tests/loadtest/ -timeout 60s -count=1 2>&1

echo ""
echo "=== Load Test Complete ==="
