#!/bin/sh
set -euo pipefail

echo "[run.sh] Starting service"

# echo "[run.sh] Running DB migrations"
# goose -dir ./db/migrations postgres "${DATABASE_URL}" up
# NB: migrations are launched from code; uncomment if needed

echo "[run.sh] Starting Go app"
exec /app/bin/app