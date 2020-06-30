#!/bin/bash

# we want to have some checks done for undefined variables
set -u

source "scripts/textutils.sh"

if [ "${HTTP_PROXY+x}" != "" ]; then
	export DOCKER_BUILD_ARGS="--build-arg http_proxy='${http_proxy}' --build-arg https_proxy='${https_proxy}' --build-arg HTTP_PROXY='${HTTP_PROXY}' --build-arg HTTPS_PROXY='${HTTPS_PROXY}' --build-arg NO_PROXY='localhost,127.0.0.1'"
	export DOCKER_RUN_ARGS="--env http_proxy='${http_proxy}' --env https_proxy='${https_proxy}' --env HTTP_PROXY='${HTTP_PROXY}' --env HTTPS_PROXY='${HTTPS_PROXY}' --env NO_PROXY='localhost,127.0.0.1'"
	export AWS_CLI_PROXY="export http_proxy='${http_proxy}'; export https_proxy='${https_proxy}'; export HTTP_PROXY='${HTTP_PROXY}'; export HTTPS_PROXY='${HTTPS_PROXY}'; export NO_PROXY='localhost,127.0.0.1';"
else
	export DOCKER_BUILD_ARGS=""
	export DOCKER_RUN_ARGS=""
	export AWS_CLI_PROXY=""
fi

# Build GO serf handlers
msg="Building GO Serf handlers, this can take a few minutes..."
printBanner "$msg
logMsg "$msg
./scripts/buildSerfHandlers.sh

# Build Retail WorkLoad Orchestrator
msg="Building RWO, this can take a few minutes..."
printBanner "$msg
logMsg "$msg
source "scripts/buildRWO.sh"

msg="Building Complete"
printBanner "$msg"
