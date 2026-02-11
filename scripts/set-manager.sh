#!/bin/bash

# Set P-Manager information on the P-Node to start automatic communication.

# Get MANAGER_URL and MANAGER_TOKEN from environment variables (passed by Makefile)
MANAGER_URL="$1"
MANAGER_TOKEN="$2"

# Read database values once
DB_PATH="$(realpath "$(dirname "$0")/../storage/database/data.json")"
NODE_TOKEN=$(jq -r '.settings.http_token' "$DB_PATH")
NODE_PORT=$(jq -r '.settings.http_port' "$DB_PATH")

# Make a local HTTP request to configure P-Node
echo "Setting P-Manager '$MANAGER_URL' on the local P-Node..."
if curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $NODE_TOKEN" \
    -d "{\"url\":\"$MANAGER_URL\",\"token\":\"$MANAGER_TOKEN\"}" \
    "http://localhost:$NODE_PORT/manager"; then
    echo "P-Node updated successfully."
else
    echo "Failed to update P-Node."
    exit 1
fi
