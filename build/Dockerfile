###########################################################################
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
# imitations under the License.
###########################################################################
FROM quay.io/operator-framework/helm-operator:v0.8.0

ARG VERSION
ARG BUILD_DATE

LABEL name="Application Navigator" \
      vendor="kAppNav" \
      version=$VERSION \
      release=$VERSION \
      created=$BUILD_DATE \
      summary="Operator image for Application Navigator" \
      description="This image contains the operator for Application Navigator"

COPY licenses/ /licenses/
COPY watches.yaml ${HOME}/watches.yaml
COPY helm-chart/ ${HOME}/helm-charts/
