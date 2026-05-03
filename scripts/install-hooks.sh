#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
git config core.hooksPath .githooks
echo "Git hooks installed from .githooks"
