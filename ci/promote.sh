#!/bin/sh -ex

# Promote the built snap to stable on the Syncloud store.
# Usage: ./ci/promote.sh <arch> <app> <deployer-base-url>

ARCH=$1
APP=$2
DEPLOYER=$3

apt update && apt install -y wget
wget ${DEPLOYER}-${ARCH} -O release --progress=dot:giga
chmod +x release
./release promote -n ${APP} -a $(dpkg --print-architecture)
