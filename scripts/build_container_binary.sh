#!/bin/sh

source ./scripts/version.sh

# build binary image
GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o iperf3-handler -ldflags "$LD_FLAGS" cmd/main.go
