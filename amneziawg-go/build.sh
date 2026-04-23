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

# Log which select case fires in main.go so we can diagnose early shutdowns.
python3 - <<'PY'
import re
p = 'main.go'
s = open(p).read()
old = '''	select {
	case <-term:
	case <-errs:
	case <-device.Wait():
	}'''
new = '''	select {
	case sig := <-term:
		logger.Errorf("Shutting down: signal %v", sig)
	case e := <-errs:
		logger.Errorf("Shutting down: errs channel: %v", e)
	case <-device.Wait():
		logger.Errorf("Shutting down: device.Wait() fired")
	}'''
if old not in s:
    raise SystemExit('select block not found')
open(p, 'w').write(s.replace(old, new))
PY

CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X github.com/amnezia-vpn/amneziawg-go/ipc.socketDirectory=/var/snap/amneziawg/current/run/amneziawg" \
  -o ${BUILD_DIR}/amneziawg-go
