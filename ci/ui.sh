#!/bin/sh -ex

PROJECT=$1
APP=$2
DISTRO=$3
VERSION=$4

getent hosts ${APP}.${DISTRO}.com | sed "s/${APP}.${DISTRO}.com/auth.${DISTRO}.com/g" | tee -a /etc/hosts

cd web
npm ci
PLAYWRIGHT_DOMAIN=${DISTRO}.com \
PLAYWRIGHT_APP=${APP} \
PLAYWRIGHT_VERSION=${VERSION} \
  npx playwright test --project=${PROJECT}
