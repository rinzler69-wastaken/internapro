#!/usr/bin/env bash
set -euo pipefail

# Deployment start script for Railway
# Builds the Svelte frontend, wires env vars to the Go API, then runs the API
# which also serves the built SPA (via newSPAHandler in backend/internal/routes).

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Railway exposes the public port via $PORT. Keep API and static on the same port.
PORT="${PORT:-8080}"

# Map Railway MySQL variables to the Go app's expected names if custom values are not provided.
export DB_HOST="${DB_HOST:-${MYSQLHOST:-localhost}}"
export DB_PORT="${DB_PORT:-${MYSQLPORT:-3306}}"
export DB_USER="${DB_USER:-${MYSQLUSER:-root}}"
export DB_PASSWORD="${DB_PASSWORD:-${MYSQLPASSWORD:-}}"
export DB_NAME="${DB_NAME:-${MYSQLDATABASE:-interna_db}}"

# Server configuration
export SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
export SERVER_PORT="${SERVER_PORT:-$PORT}"
export APP_ENV="${APP_ENV:-production}"

# Point backend SPA handler to the built frontend output
export FRONTEND_DIST_DIR="${FRONTEND_DIST_DIR:-$ROOT/frontend/dist}"

echo "==> Preparing frontend build"
if [ ! -d "$ROOT/frontend/node_modules" ]; then
  npm ci --prefix "$ROOT/frontend"
fi
npm run build --prefix "$ROOT/frontend"

echo "==> Ensuring upload directories exist"
mkdir -p "$ROOT/backend/uploads/tasks" "$ROOT/backend/uploads/leaves"

echo "==> Building backend"
mkdir -p "$ROOT/bin"
cd "$ROOT/backend"
GOFLAGS="${GOFLAGS:-}"
CGO_ENABLED="${CGO_ENABLED:-0}" go build $GOFLAGS -o "$ROOT/bin/server" ./cmd/server

echo "==> Launching server on port $SERVER_PORT"
cd "$ROOT"
exec "$ROOT/bin/server"
