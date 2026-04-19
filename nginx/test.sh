#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
BUILD_DIR=${DIR}/../build/snap/nginx

# Sanity check: the wrapper can run nginx -v against the bundled binary.
${BUILD_DIR}/bin/nginx.sh -v
