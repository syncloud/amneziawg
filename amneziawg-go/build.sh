#!/bin/sh -ex

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

wget -q \
  https://github.com/amnezia-vpn/amneziawg-go/archive/refs/tags/v${VERSION}.tar.gz \
  -O src.tar.gz
tar xf src.tar.gz --strip-components=1 --no-same-owner --no-same-permissions

# The inotify watcher in ipc/uapi_linux.go reads events into a zero-byte
# buffer (`var buf [0]byte`). Linux's inotify_read returns EINVAL if the
# buffer is too small for the next event, so the first IN_ATTRIB fires
# and the daemon silently shuts down. Size the buffer properly.
python3 - <<'PY'
p = 'ipc/uapi_linux.go'
s = open(p).read()
old = 'var buf [0]byte'
new = 'var buf [4096]byte'
if old not in s:
    raise SystemExit('inotify buffer decl not found in ' + p)
open(p, 'w').write(s.replace(old, new))
PY

CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X github.com/amnezia-vpn/amneziawg-go/ipc.socketDirectory=/var/snap/amneziawg/current/run/amneziawg" \
  -o ${BUILD_DIR}/amneziawg-go
