#!/usr/bin/env bash
set -euo pipefail

ARCH="${GOARCH:-amd64}"
SERVER_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MAIN="$SERVER_ROOT/."
GIT_DATE="$(git show -s --date=format:%Y%m%d --format=%cd HEAD)"
GIT_HASH="$(git rev-parse --short=8 HEAD)"
VERSION="${GIT_DATE}-${GIT_HASH}"
LDFLAGS="-s -w -X main.Version=${VERSION}"

GOOS=linux GOARCH="$ARCH" go build -ldflags="$LDFLAGS" -o "$SERVER_ROOT/mkp-server" "$MAIN"
echo "Built $SERVER_ROOT/mkp-server"
