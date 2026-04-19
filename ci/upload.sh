#!/bin/sh -ex

# Upload the built snap to the Syncloud store.
# Usage: ./ci/upload.sh <arch> <deployer-base-url> <branch>

ARCH=$1
DEPLOYER=$2
BRANCH=$3

PACKAGE=$(cat package.name)
apt update && apt install -y wget
wget ${DEPLOYER}-${ARCH} -O release --progress=dot:giga
chmod +x release
./release publish -f ${PACKAGE} -b ${BRANCH}
