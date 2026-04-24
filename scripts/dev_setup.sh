#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

go mod tidy

go run ./cmd/gogal init .

echo "Dev setup complete. Next: go run ./cmd/gogal doctor"
