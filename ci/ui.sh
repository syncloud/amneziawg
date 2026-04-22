#!/bin/sh -ex

PROJECT=$1
APP=$2
DISTRO=$3
VERSION=$4

getent hosts ${APP}.${DISTRO}.com | sed "s/${APP}.${DISTRO}.com/auth.${DISTRO}.com/g" | tee -a /etc/hosts

ART=/drone/src/artifact/${PROJECT}
mkdir -p "$ART"
trap 'cp -r /drone/src/web/test-results "$ART/" 2>/dev/null; cp -r /drone/src/web/playwright-report "$ART/" 2>/dev/null; chmod -R a+r "$ART" 2>/dev/null; exit' EXIT INT TERM

cd web
npm ci
PLAYWRIGHT_DOMAIN=${DISTRO}.com \
PLAYWRIGHT_APP=${APP} \
PLAYWRIGHT_VERSION=${VERSION} \
  npx playwright test --project=${PROJECT}
