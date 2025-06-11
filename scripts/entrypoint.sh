#!/bin/sh

echo "[INFO] Starting entrypoint script.."

echo "Starting app with config:"
echo "INIT_SQLITE=$INIT_SQLITE"
echo "SQLITE_PATH=$SQLITE_PATH"

if [ "$INIT_SQLITE" = "true" ]; then
  if [ ! -f "$SQLITE_PATH" ]; then
    echo "[INFO] Initializing new SQLite DB at $SQLITE_PATH..."
    ./db-init -db "$SQLITE_PATH"
  else
    echo "[INFO] SQLite DB already exists at $SQLITE_PATH. Skipping init."
  fi
else
  echo "[INFO] INIT_SQLITE is false, skipping Default DB init."
fi

echo "[INFO] Launching server..."
exec /server