#!/bin/bash

source ./scripts/version.sh

# container registry
REGISTRY='quay.io/jkandasa/iperf3-handler'
PLATFORMS="linux/386,linux/amd64,linux/arm64,linux/ppc64le,linux/s390x"
IMAGE_TAG=${VERSION}

# build and push to quay.io
docker buildx build --push \
  --progress=plain \
  --build-arg=GOPROXY=${GOPROXY} \
  --platform ${PLATFORMS} \
  --file Dockerfile \
  --tag ${REGISTRY}:${IMAGE_TAG} .
