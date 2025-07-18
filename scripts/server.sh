#!/bin/bash

# StackAgent Server Launch Script
# This script ensures proper signal handling for graceful shutdown

set -e

echo "🚀 Starting StackAgent Server..."
echo "💡 Use Ctrl+C for graceful shutdown"
echo

# Build the server first
go build -o bin/stackagent-server cmd/stackagent-server/main.go

# Run the server with proper signal handling
exec ./bin/stackagent-server 