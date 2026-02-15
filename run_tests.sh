#!/bin/bash

echo "=== Running Unit Tests ==="
echo ""

echo "📦 Testing API Layer..."
go test -v ./cmd/api -run "Test.*" 2>&1 | grep -E "PASS|FAIL|RUN|ok|FAIL"
echo ""

echo "📦 Testing Application Layer..."
go test -v ./internal/application -run "Test.*" 2>&1 | grep -E "PASS|FAIL|RUN|ok|FAIL"
echo ""

echo "📦 Testing Worker..."
go test -v ./internal/worker -run "Test.*" 2>&1 | grep -E "PASS|FAIL|RUN|ok|FAIL"
echo ""

echo "=== Coverage Report ==="
go test -cover ./cmd/api ./internal/application ./internal/worker
echo ""

echo "=== Generating HTML Coverage Report ==="
go test -coverprofile=coverage.out ./cmd/api ./internal/application ./internal/worker
go tool cover -html=coverage.out -o coverage.html
echo "✅ Coverage report generated: coverage.html"
