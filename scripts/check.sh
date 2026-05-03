#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> gofmt"
unformatted="$(gofmt -l cmd internal)"
if [[ -n "$unformatted" ]]; then
  echo "$unformatted"
  echo "Run: gofmt -w cmd internal"
  exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test ./...

if command -v docker >/dev/null 2>&1; then
  echo "==> docker compose config"
  export GUNOI_IMAGE="${GUNOI_IMAGE:-ghcr.io/baditaflorin/calendar-ridicare-gunoi-arad:latest}"
  export GUNOI_PUBLIC_BASE_URL="${GUNOI_PUBLIC_BASE_URL:-http://localhost:26453}"
  export GUNOI_REFRESH_INTERVAL="${GUNOI_REFRESH_INTERVAL:-6h}"
  docker compose config >/dev/null
fi
