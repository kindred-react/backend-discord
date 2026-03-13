#!/bin/bash

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

echo "=========================================="
echo "Starting Backend Discord Server"
echo "=========================================="

SERVER_PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8080")
echo "Server Port: $SERVER_PORT"

echo ""
echo "[1/4] Killing existing processes on port $SERVER_PORT..."

PID=$(lsof -ti :$SERVER_PORT 2>/dev/null || true)
if [ -n "$PID" ]; then
    echo "Found process $PID running on port $SERVER_PORT"
    kill -9 $PID 2>/dev/null || true
    echo "Killed process $PID"
    sleep 1
else
    echo "No process found on port $SERVER_PORT"
fi

echo ""
echo "[2/4] Clearing cache..."

if [ -d "bin" ]; then
    echo "Removing bin/ directory..."
    rm -rf bin/
fi

if [ -f "server.bin" ]; then
    echo "Removing server.bin..."
    rm -f server.bin
fi

if [ -f "server_clean" ]; then
    echo "Removing server_clean..."
    rm -f server_clean
fi

if [ -f "server_final" ]; then
    echo "Removing server_final..."
    rm -f server_final
fi

if [ -f "server_new" ]; then
    echo "Removing server_new..."
    rm -f server_new
fi

if [ -f "server.log" ]; then
    echo "Removing server.log..."
    rm -f server.log
fi

echo "Running go clean..."
go clean -cache -modcache -testcache 2>/dev/null || true

echo ""
echo "[3/4] Building server..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

go build -o bin/server ./cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "Error: Build failed"
    exit 1
fi

echo "Build successful: bin/server"

echo ""
echo "[4/4] Starting server..."

if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

./bin/server
