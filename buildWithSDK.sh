#!/bin/bash

# Copyright 2019 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
