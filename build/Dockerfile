###########################################################################
# Copyright 2019, 2020 IBM Corporation
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

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ARG VERSION
ARG BUILD_DATE
ARG PROJECT_NAME

LABEL name="Application Navigator" \
      vendor="kAppNav" \
      version=$VERSION \
      release=$VERSION \
      created=$BUILD_DATE \
      summary="Operator image for Application Navigator" \
      description="This image contains the operator for Application Navigator"

ENV OPERATOR=/usr/local/bin/kappnav-operator \
    USER_UID=1001 \
    USER_NAME=kappnav-operator

# install operator binary
COPY build/_output/bin/${PROJECT_NAME} ${OPERATOR}

# copying various resources into the image
COPY deploy/default_values.yaml deploy/
COPY deploy/maps/ maps/
COPY deploy/crds/extensions crds/

# get application CRD from Kubernetes Application SIG
RUN curl -fsSLO --compressed https://raw.githubusercontent.com/kubernetes-sigs/application/master/config/crd/bases/app.k8s.io_applications.yaml \
    && mv app.k8s.io_applications.yaml crds/

# copy license files into the image
COPY licenses/ /licenses/

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
