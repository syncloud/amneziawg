#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
BUILD_DIR=${DIR}/../build/snap/amneziawg-tools

${BUILD_DIR}/bin/awg --version
${BUILD_DIR}/bin/awg-quick --help 2>&1 | head -1 || true
