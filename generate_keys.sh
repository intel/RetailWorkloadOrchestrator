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

msg="Generating SSL Keys..."
printBanner $msg
logMsg $msg
mkdir -p node_ca_key
cd node_ca_key
../scripts/credtool_create_ca.sh rwo
cd -
scripts/credtool_create_keys.sh rwo.key rwo.crt

msg="Generating Cluster Key..."
printBanner $msg
logMsg $msg
mkdir -p node_keys/rwo
param_rwokey="$(tr </dev/urandom -dc +a-fA-F0-9 | head -c43)="
echo '[ \"${param_rwokey}\" ]' > node_keys/rwo/keyring.json

msg="Root CA Key is in ./node_ca_key and the Keys for each node are in node_keys.  These should copied to /etc/ssl for each node.  For more details, please refer to https://github.com/intel/RetailWorkloadOrchestrator/blob/master/docs/02_Security.md"
printBanner $msg