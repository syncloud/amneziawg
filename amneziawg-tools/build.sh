#!/bin/sh -ex

# Build the AmneziaWG userspace tools (awg + awg-quick).
# Usage: ./amneziawg-tools/build.sh <version>
# Version is pinned in the repo's .drone.jsonnet.
#
# awg is a small C binary; awg-quick is a bash script.
# Build statically (musl, or glibc with -static) so the resulting
# binary has no runtime lib dependency.

if [ -z "$1" ]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

VERSION=$1
DIR=$( cd "$( dirname "$0" )" && pwd )
BUILD_DIR=${DIR}/../build/snap/amneziawg-tools
mkdir -p ${BUILD_DIR}/bin

WORK=${DIR}/../build/amneziawg-tools-src
rm -rf ${WORK}
mkdir -p ${WORK}
cd ${WORK}

wget --progress=dot:giga \
  https://github.com/amnezia-vpn/amneziawg-tools/archive/refs/tags/v${VERSION}.tar.gz \
  -O src.tar.gz
tar xf src.tar.gz --strip-components=1

cd src
make LDFLAGS='-static -s' awg
cp awg ${BUILD_DIR}/bin/awg

# awg-quick is a bash script; ship it verbatim.
cp wg-quick/linux.bash ${BUILD_DIR}/bin/awg-quick
chmod +x ${BUILD_DIR}/bin/awg-quick

${BUILD_DIR}/bin/awg --version || true
