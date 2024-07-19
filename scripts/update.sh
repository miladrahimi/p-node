#!/bin/bash

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root"
    exit 1
fi

# TODO: Remote this
SETUP_SCRIPT="$(dirname "$0")/setup.sh"
$SETUP_SCRIPT

# Store update time
rm -f ./storage/logs/*.log
date '+%Y-%m-%d %H:%M:%S' > ./storage/app/update.txt
