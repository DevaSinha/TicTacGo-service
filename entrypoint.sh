#!/bin/sh

if [ -n "$DATABASE_URL" ]; then
  DB_ADDR=$(echo $DATABASE_URL | sed -e 's/^postgres:\/\///' -e 's/^postgresql:\/\///')
else
  echo "DATABASE_URL is not set"
  exit 1
fi

# Use NAKAMA_SERVER_KEY from env, fallback to default if not set
S_KEY="${NAKAMA_SERVER_KEY:-tictactoe-server-key}"

echo "Starting Nakama with provided database and server key..."

# Run migrations
/nakama/nakama migrate up --database.address "$DB_ADDR"

# Start Nakama with explicit flags (Flags always override local.yml)
exec /nakama/nakama --config /nakama/data/local.yml \
    --database.address "$DB_ADDR" \
    --socket.server_key "$S_KEY"
