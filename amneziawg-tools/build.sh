#!/bin/sh -ex

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

wget --tries=5 --waitretry=10 --timeout=60 \
  https://github.com/amnezia-vpn/amneziawg-tools/archive/refs/tags/v${VERSION}.tar.gz \
  -O src.tar.gz
tar xf src.tar.gz --strip-components=1 --no-same-owner --no-same-permissions

cd src
make RUNSTATEDIR=/var/snap/amneziawg/current/run LDFLAGS='-static -s' wg
cp wg ${BUILD_DIR}/bin/awg

cp wg-quick/linux.bash ${BUILD_DIR}/bin/awg-quick
chmod +x ${BUILD_DIR}/bin/awg-quick

${BUILD_DIR}/bin/awg --version || true
