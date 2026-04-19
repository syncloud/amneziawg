#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/../build/snap/nginx
mkdir -p ${BUILD_DIR}

# Run from the `nginx:*-alpine` image in CI — copy the alpine rootfs
# pieces we need to run nginx via its own ld-musl. Same approach as
# owncast/nginx/build.sh.
cp -r /etc ${BUILD_DIR}
cp -r /opt ${BUILD_DIR}
cp -r /usr ${BUILD_DIR}
cp -r /bin ${BUILD_DIR}
cp -r /lib ${BUILD_DIR}

cp -r ${DIR}/bin/* ${BUILD_DIR}/bin
