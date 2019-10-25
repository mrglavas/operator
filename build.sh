#!/bin/bash
IMAGE=kappnav-operator
VERSION=0.1.1

echo "Building ${IMAGE} ${VERSION}"

docker build --pull --build-arg VERSION=${VERSION} --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')  -f build/Dockerfile -t ${IMAGE} . 
