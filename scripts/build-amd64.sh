#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

image="${IMAGE:-ghcr.io/baditaflorin/calendar-ridicare-gunoi-arad}"
tag="${TAG:-$(git rev-parse --short HEAD)}"
output="--load"

if [[ "${PUSH:-false}" == "true" ]]; then
  output="--push"
fi

docker buildx build \
  --platform linux/amd64 \
  -t "$image:$tag" \
  -t "$image:latest" \
  $output \
  .

echo "Built $image:$tag and $image:latest for linux/amd64"
