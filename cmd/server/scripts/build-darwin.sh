#!/usr/bin/env bash
set -euo pipefail

ARCH="${GOARCH:-amd64}"
SERVER_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MAIN="$SERVER_ROOT/main.go"

GOOS=darwin GOARCH="$ARCH" go build -ldflags="-s -w" -o "$SERVER_ROOT/mkp-server" "$MAIN"
echo "Built $SERVER_ROOT/mkp-server"
