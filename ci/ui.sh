#!/bin/sh -ex

PROJECT=$1
APP=$2
DISTRO=$3
VERSION=$4

cd web
npm ci
PLAYWRIGHT_DOMAIN=${DISTRO}.com \
PLAYWRIGHT_APP=${APP} \
PLAYWRIGHT_VERSION=${VERSION} \
  npx playwright test --project=${PROJECT}
