#!/bin/bash
IMAGE=kappnav-operator

echo build $IMAGE ...

echo "Building ${IMAGE}"

docker build -f build/Dockerfile -t ${IMAGE} . 
