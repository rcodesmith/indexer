#!/usr/bin/env sh
set -e

echo "Starting indexer, connecting to algod node.  Using DB $DATABASE_NAME"

# Start indexer, connecting to node & DB.  Get node API token from shared volume
./cmd/algorand-indexer/algorand-indexer daemon \
  -P "host=indexer-db port=5432 user=${DATABASE_USER} password=${DATABASE_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable" \
  --algod-net="http://${ALGORAND_HOST}:8080" --algod-token="$(cat /var/algorand/data/algod.token)"
