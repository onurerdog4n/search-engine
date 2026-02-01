#!/bin/bash

# Dependency Audit Script
# This script checks for outdated dependencies and vulnerabilities

echo "🔍 Checking Go dependencies..."
echo ""

# Check for updates
echo "📦 Available updates:"
go list -u -m all 2>/dev/null || echo "Error: go command not found"
echo ""

# Tidy and verify
echo "🧹 Tidying dependencies..."
go mod tidy 2>/dev/null || echo "Error: go command not found"
echo ""

echo "✅ Verifying dependencies..."
go mod verify 2>/dev/null || echo "Error: go command not found"
echo ""

# Check for known vulnerabilities (requires govulncheck)
echo "🔒 Checking for vulnerabilities..."
if command -v govulncheck &> /dev/null; then
    govulncheck ./...
else
    echo "⚠️  govulncheck not installed. Install with:"
    echo "   go install golang.org/x/vuln/cmd/govulncheck@latest"
fi
echo ""

# Summary
echo "✨ Audit complete!"
echo ""
echo "Next steps:"
echo "1. Review outdated dependencies"
echo "2. Update dependencies: go get -u <package>"
echo "3. Run tests: go test ./..."
echo "4. Commit changes: git add go.mod go.sum && git commit"
