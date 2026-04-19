#!/bin/sh -ex

# Upgrade-path test runner (amd64 only). Invoked by drone.
# Usage: ./ci/upgrade.sh <app> <distro> <build-number> <browser>

APP=$1
DISTRO=$2
VERSION=$3
BROWSER=$4

cd test
./deps.sh
py.test -x -s upgrade.py --distro=${DISTRO} --ver=${VERSION} --app=${APP} --browser=${BROWSER}
