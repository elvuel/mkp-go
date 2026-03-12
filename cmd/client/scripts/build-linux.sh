#!/usr/bin/env bash
set -euo pipefail

ARCH="${GOARCH:-amd64}"
CLIENT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILL_SCRIPTS="$CLIENT_ROOT/openclaw_skills/mkp/scripts"
MAIN="$CLIENT_ROOT/."
GIT_DATE="$(git show -s --date=format:%Y%m%d --format=%cd HEAD)"
GIT_HASH="$(git rev-parse --short=8 HEAD)"
VERSION="${GIT_DATE}-${GIT_HASH}"
LDFLAGS="-s -w -X main.Version=${VERSION}"

mkdir -p "$SKILL_SCRIPTS"
GOOS=linux GOARCH="$ARCH" go build -ldflags="$LDFLAGS" -o "$SKILL_SCRIPTS/mkp" "$MAIN"
echo "Built $SKILL_SCRIPTS/mkp"
