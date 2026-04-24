#!/bin/sh -ex

APP=$1
DISTRO=$2
VERSION=$3

cd test
./deps.sh
py.test -x -s upgrade.py --distro=${DISTRO} --ver=${VERSION} --app=${APP}
