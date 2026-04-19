#!/bin/sh -ex

# Integration-test runner. Invoked by drone per-distro.
# Usage: ./ci/integration.sh <app> <distro> <build-number>

APP=$1
DISTRO=$2
VERSION=$3

cd test
./deps.sh
py.test -x -s test.py --distro=${DISTRO} --ver=${VERSION} --app=${APP}
