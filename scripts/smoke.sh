#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

tmp_dir="$(mktemp -d)"
port="${PORT:-18081}"
pid=""

cleanup() {
  if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
    kill "$pid" >/dev/null 2>&1 || true
    wait "$pid" >/dev/null 2>&1 || true
  fi
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

echo "==> build smoke binary"
go build -o "$tmp_dir/gunoiarad" ./cmd/gunoiarad

echo "==> run live ETL smoke"
"$tmp_dir/gunoiarad" etl --db "$tmp_dir/gunoiarad.db" --raw-dir "$tmp_dir/raw"

echo "==> start smoke server"
"$tmp_dir/gunoiarad" serve \
  --addr "127.0.0.1:$port" \
  --db "$tmp_dir/gunoiarad.db" \
  --raw-dir "$tmp_dir/raw" \
  --public-base-url "http://127.0.0.1:$port" &
pid="$!"

for _ in {1..40}; do
  if curl -fsS "http://127.0.0.1:$port/readyz" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -fsS "http://127.0.0.1:$port/readyz" >/dev/null
curl -fsS "http://127.0.0.1:$port/api/neighborhoods" | grep -q '"norm":"muresel"'

search_json="$(curl -fsS "http://127.0.0.1:$port/api/places?cartier_norm=muresel&q=densu&limit=1")"
echo "$search_json" | grep -q 'Nicolae Densu'
place_id="$(printf '%s' "$search_json" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' | head -1)"
if [[ -z "$place_id" ]]; then
  echo "could not find place id in search response"
  echo "$search_json"
  exit 1
fi

curl -fsS "http://127.0.0.1:$port/api/events?place_id=$place_id&from=2026-05-01&to=2026-05-31" | grep -q '"waste_type":"plastic_metal"'
curl -fsS "http://127.0.0.1:$port/ics?place_id=$place_id" | grep -q 'BEGIN:VCALENDAR'
curl -fsS "http://127.0.0.1:$port/print?place_id=$place_id&month=2026-05" | grep -q 'Calendar de colectare'
curl -fsS "http://127.0.0.1:$port/metrics" | grep -q 'gunoi_arad_http_requests_total'

echo "Smoke OK on port $port"
