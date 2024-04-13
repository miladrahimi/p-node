#!/bin/bash

DB_PATH="$(realpath "$(dirname "$0")/../storage/database/app.json")"
if [ -f "$DB_PATH" ]; then
  printf "IP: " && curl ifconfig.io
  printf "DB: " && cat "$DB_PATH" && printf "\n"
else
  echo "The app is not ready yet. Please try again..."
fi
