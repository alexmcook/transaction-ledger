#!/usr/bin/env bash
set -euo pipefail

if [ -f .env ]; then
  set -o allexport
  source .env
fi

MIGRATE_CMD="${MIGRATE_CMD:-migrate}"
MIGRATIONS_PATH="${MIGRATIONS_PATH:-$(pwd)/migrations}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 {up|down}"
  exit 2
fi

cmd="$1"

if ! command -v "$MIGRATE_CMD" >/dev/null 2>&1; then
  echo "migrate CLI not found in PATH"
  exit 1
fi

case "$cmd" in
  up)
    "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
    ;;
  down)
    "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down
    ;;
  *)
    echo "Usage: $0 {up|down}"
    exit 2
    ;;
esac
