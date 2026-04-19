#!/bin/sh -ex

# Playwright UI test runner. Invoked by drone.
# Usage: ./ci/ui.sh <desktop|mobile> <app> <distro> <build-number>

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
