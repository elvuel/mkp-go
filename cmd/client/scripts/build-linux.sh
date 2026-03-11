#!/usr/bin/env bash
set -euo pipefail

ARCH="${GOARCH:-amd64}"
CLIENT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILL_SCRIPTS="$CLIENT_ROOT/openclaw_skills/mkp/scripts"
MAIN="$CLIENT_ROOT/main.go"

mkdir -p "$SKILL_SCRIPTS"
GOOS=linux GOARCH="$ARCH" go build -ldflags="-s -w" -o "$SKILL_SCRIPTS/mkp" "$MAIN"
echo "Built $SKILL_SCRIPTS/mkp"
