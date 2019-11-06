#!/bin/bash
. ../build/version.sh
IMAGE=kappnav-operator

# Update version numbers in yaml files
./updateVersionsinYamls.sh

echo "Building ${IMAGE} ${VERSION}"
docker build --pull --build-arg VERSION=${VERSION} --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')  -f build/Dockerfile -t ${IMAGE} .

# Restore original yaml files
./restoreYamls.sh
