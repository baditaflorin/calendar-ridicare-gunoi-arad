#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

db="${GUNOI_DB_PATH:-.smoke/static-export.db}"
raw_dir="${GUNOI_RAW_DIR:-.smoke/raw}"
out="${OUT:-docs}"

mkdir -p "$(dirname "$db")" "$raw_dir"

if [[ "${SKIP_ETL:-false}" != "true" ]]; then
  go run ./cmd/gunoiarad etl --db "$db" --raw-dir "$raw_dir"
fi

go run ./cmd/gunoiarad export-static --db "$db" --out "$out"

echo "Static GitHub Pages bundle written to $out"
