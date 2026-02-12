#!/bin/bash

# Reveal required information about the node for P-Manager.

DB_PATH="$(realpath "$(dirname "$0")/../storage/database/data.json")"

get_ssh_port() {
  local sshd_bin=""
  local ports=""

  if command -v sshd >/dev/null 2>&1; then
    sshd_bin="sshd"
  elif [ -x /usr/sbin/sshd ]; then
    sshd_bin="/usr/sbin/sshd"
  fi

  if [ -n "$sshd_bin" ]; then
    ports="$($sshd_bin -T 2>/dev/null | awk '$1 == "port" {print $2}' | paste -sd, -)"
  fi

  if [ -z "$ports" ] && [ -f /etc/ssh/sshd_config ]; then
    ports="$(awk 'BEGIN {IGNORECASE=1} $1 == "Port" {print $2}' /etc/ssh/sshd_config 2>/dev/null | paste -sd, -)"
  fi

  if [ -z "$ports" ]; then
    ports="22"
  fi

  printf "%s" "$ports"
}

get_http_port() {
  local port=""

  if command -v jq >/dev/null 2>&1; then
    port="$(jq -r '.settings.http_port // empty' "$DB_PATH" 2>/dev/null)"
  fi

  if [ -z "$port" ]; then
    port="$(sed -n 's/.*"http_port"[[:space:]]*:[[:space:]]*\\([0-9][0-9]*\\).*/\\1/p' "$DB_PATH" 2>/dev/null | head -n1)"
  fi

  printf "%s" "$port"
}

get_http_token() {
  local token=""

  if command -v jq >/dev/null 2>&1; then
    token="$(jq -r '.settings.http_token // empty' "$DB_PATH" 2>/dev/null)"
  fi

  if [ -z "$token" ]; then
    token="$(sed -n 's/.*"http_token"[[:space:]]*:[[:space:]]*"\\([^"]*\\)".*/\\1/p' "$DB_PATH" 2>/dev/null | head -n1)"
  fi

  printf "%s" "$token"
}

json_string() {
  local value="$1"

  if [ -z "$value" ]; then
    printf "null"
  else
    printf "\"%s\"" "$(printf '%s' "$value" | sed 's/\\/\\\\/g; s/"/\\"/g')"
  fi
}

json_number() {
  local value="$1"

  if [[ -n "$value" && "$value" =~ ^[0-9]+$ ]]; then
    printf "%s" "$value"
  else
    printf "null"
  fi
}

if [ -f "$DB_PATH" ]; then
  ip="$(curl -s --max-time 5 ifconfig.io 2>/dev/null)"
  http_port="$(get_http_port)"
  http_token="$(get_http_token)"
  ssh_port="$(get_ssh_port)"
  ssh_user="$(id -un 2>/dev/null || whoami)"

  if [ -z "$http_port" ]; then
    message="HTTP port not found in database."
  elif [ -z "$http_token" ]; then
    message="HTTP token not found in database."
  elif [ -z "$ip" ]; then
    message="Public IP is not available."
  else
    message="Copy the JSON below and paste in the P-Manager input:"
  fi
else
  ip=""
  http_port=""
  http_token=""
  ssh_port=""
  ssh_user=""
  message="The app is not ready yet. Please try again..."
fi

printf "%s\n" "$message"
printf "{"
printf "\"ip\":%s," "$(json_string "$ip")"
printf "\"http_port\":%s," "$(json_number "$http_port")"
printf "\"http_token\":%s," "$(json_string "$http_token")"
printf "\"ssh_port\":%s," "$(json_string "$ssh_port")"
printf "\"ssh_user\":%s" "$(json_string "$ssh_user")"
printf "}\n"
