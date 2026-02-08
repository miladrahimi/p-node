#!/bin/bash

# Set P-Manager configs on the node to start communicating with P-Manager.

DB_PATH="$(realpath "$(dirname "$0")/../storage/database/data.json")"

# Get URL and TOKEN from environment variables (passed by Makefile)
URL="$1"
TOKEN="$2"

# Read database values once
HTTP_TOKEN=$(jq -r '.settings.http_token' "$DB_PATH")
HTTP_PORT=$(jq -r '.settings.http_port' "$DB_PATH")

# Make the HTTP request
echo "Setting P-Manager with URL: $URL"
if curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $HTTP_TOKEN" \
    -d "{\"url\":\"$URL\",\"token\":\"$TOKEN\"}" \
    "http://localhost:$HTTP_PORT/manager"; then
    echo "Manager configs updated successfully"
else
    echo "Failed to update manager configs"
    exit 1
fi
