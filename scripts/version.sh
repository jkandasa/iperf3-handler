#!/bin/bash

# this script updateds version information

# version details
export BUILD_DATE=`date -u +'%Y-%m-%dT%H:%M:%S%:z'`
export GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
export GIT_SHA=`git rev-parse HEAD`
export GIT_SHA_SHORT=`git rev-parse --short HEAD`
export VERSION_PKG="github.com/jkandasa/iperf3-handler/pkg/version"

# update tag, if available
if [ ${GIT_BRANCH} = "HEAD" ]; then
  export GIT_BRANCH=`git describe --abbrev=0 --tags`
fi

# update version number
export VERSION=`echo ${GIT_BRANCH} |  awk 'match($0, /([0-9]*\.[0-9]*\.[0-9]*)$/) { print substr($0, RSTART, RLENGTH) }'`
if [ -z "$VERSION" ]; then
  # takes version from versions file and adds devel suffix with that
  export STATIC_VERSION=`grep "iperf3-handler=" versions.txt | awk -F= '{print $2}'`
  export VERSION="${STATIC_VERSION}-devel"
fi

export LD_FLAGS="-X $VERSION_PKG.version=$VERSION -X $VERSION_PKG.buildDate=$BUILD_DATE -X $VERSION_PKG.gitCommit=$GIT_SHA"
