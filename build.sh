#!/bin/bash
# Travis builds won't have a peer build dir
VERSION=x.x.x
if [ -e ../build/version.sh ]; then
    . ../build/version.sh
fi
IMAGE=kappnav-operator

# Update version numbers in yaml files
./updateVersionsinYamls.sh

echo "Building ${IMAGE} ${VERSION}"
docker build --pull --build-arg VERSION=${VERSION} --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')  -f build/Dockerfile -t ${IMAGE} .

# Restore original yaml files
./restoreYamls.sh
