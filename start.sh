#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. FORCE Railway's dynamic PORT into your APP's expected variable
# If Railway gives us 3456, we tell our Go app to listen on 3456
export SERVER_PORT="${PORT:-8080}"
export SERVER_HOST="0.0.0.0"

# 2. Map Database Variables (Your mapping is already great)
export DB_HOST="${DB_HOST:-${MYSQLHOST:-localhost}}"
export DB_PORT="${DB_PORT:-${MYSQLPORT:-3306}}"
export DB_USER="${DB_USER:-${MYSQLUSER:-root}}"
export DB_PASSWORD="${DB_PASSWORD:-${MYSQLPASSWORD:-}}"
export DB_NAME="${DB_NAME:-${MYSQLDATABASE:-interna_db}}"

# 3. Ensure your persistent volume directories exist
# If UPLOAD_DIR is /app/uploads, this creates the folders on your Volume
echo "==> Ensuring upload directories exist at $UPLOAD_DIR"
export UPLOAD_DIR="${UPLOAD_DIR-}" # Initialize if empty
export UPLOAD_DIR="${UPLOAD_DIR:-$ROOT/uploads}"
# We skip building here because we'll do that in the 'Build Command' stage
echo "==> Launching server on port $SERVER_PORT"
exec "$ROOT/bin/server"

# Check if the binary exists before trying to run it
if [ ! -f "$ROOT/bin/server" ]; then
    echo "ERROR: Binary not found at $ROOT/bin/server"
    exit 1
fi

echo "==> Launching server on port $SERVER_PORT"
exec "$ROOT/bin/server"