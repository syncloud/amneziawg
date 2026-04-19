#!/bin/sh -ex

# Build the userspace amneziawg-go data plane.
# Usage: ./amneziawg-go/build.sh <version>
# Version is pinned in the repo's .drone.jsonnet so bumps are visible
# as CI config diffs.

if [ -z "$1" ]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

VERSION=$1
DIR=$( cd "$( dirname "$0" )" && pwd )
BUILD_DIR=${DIR}/../build/snap/amneziawg-go
mkdir -p ${BUILD_DIR}

WORK=${DIR}/../build/amneziawg-go-src
rm -rf ${WORK}
mkdir -p ${WORK}
cd ${WORK}

wget --progress=dot:giga \
  https://github.com/amnezia-vpn/amneziawg-go/archive/refs/tags/v${VERSION}.tar.gz \
  -O src.tar.gz
tar xf src.tar.gz --strip-components=1

CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w" \
  -o ${BUILD_DIR}/amneziawg-go
