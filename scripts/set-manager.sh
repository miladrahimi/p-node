#!/bin/bash

DB_PATH="$(realpath "$(dirname "$0")/../storage/database/app.json")"

# Get URL and TOKEN from environment variables (passed by Makefile)
URL="$1"
TOKEN="$2"

# Read database values once
HTTP_TOKEN=$(jq -r '.settings.http_token' "$DB_PATH")
HTTP_PORT=$(jq -r '.settings.http_port' "$DB_PATH")

# Make the HTTP request
echo "Setting manager with URL: $URL"
if curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $HTTP_TOKEN" \
    -d "{\"url\":\"$URL\",\"token\":\"$TOKEN\"}" \
    "http://localhost:$HTTP_PORT/v1/manager"; then
    echo "Manager configs updated successfully"
else
    echo "Failed to update manager configs"
    exit 1
fi
