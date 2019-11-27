#!/bin/bash
# Travis builds won't have a peer build dir
VERSION=x.x.x
if [ -e ../build/version.sh ]; then
    . ../build/version.sh
fi
IMAGE=kappnav-operator

CURRENT=`pwd`
PROJECT=`basename "$CURRENT"`

# Update version numbers in yaml files
./updateVersionsinYamls.sh

echo "Building ${IMAGE} ${VERSION}"
operator-sdk build --image-build-args "--pull --build-arg VERSION=${VERSION} --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') --build-arg PROJECT_NAME=${PROJECT}" ${IMAGE}

# Restore original yaml files
./restoreYamls.sh
