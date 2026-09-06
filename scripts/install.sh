#!/bin/bash

# One-line installer for P-Node (Debian/Ubuntu). Run as root, e.g.:
#   curl -fsSL https://raw.githubusercontent.com/miladrahimi/p-node/master/scripts/install.sh | sudo bash
#
# It installs the requirements, clones the repo into the next free p-node-N
# directory, sets up the service, and prints the info required for P-Manager.

set -e

REPO="https://github.com/miladrahimi/p-node.git"

if [ "$(id -u)" -ne 0 ]; then
    echo "This installer must be run as root (e.g. pipe it into 'sudo bash')."
    exit 1
fi

# Install the minimum needed to clone and run make; make setup installs the rest.
if ! command -v git >/dev/null 2>&1 || ! command -v make >/dev/null 2>&1; then
    apt-get -y update
    apt-get -y install git make
fi

# Pick the next free p-node-N directory (one per instance on a server).
i=1
while [ -d "p-node-$i" ]; do
    i=$((i + 1))
done
DIR="p-node-$i"

echo "Installing P-Node into $(pwd)/$DIR ..."
git clone "$REPO" "$DIR"
cd "$DIR"

make setup
make info
