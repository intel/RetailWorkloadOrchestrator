#!/bin/sh
set -e
PLUGIN_NAME=edge/glusterfs-plugin
PLUGIN_TAG=1.0
PLUGIN_IMAGE=${PLUGIN_NAME}:${PLUGIN_TAG}

echo -e "==> create rootfs directory in ./plugin/rootfs"
mkdir -p ./plugin/rootfs
cntr=${PLUGIN_NAME}-${PLUGIN_TAG}-$(date +'%Y%m%d-%H%M%S'); \
docker create --name $$cntr ${PLUGIN_IMAGE}; \
docker export $$cntr | tar -x -C ./plugin/rootfs; \
docker rm -vf $$cntr
echo -e "### copy config.json to ./plugin/"

cp ${RWO_BASE_PATH}/gluster/config.json ./plugin/
# Requires the env variables TLS_KEY, TLS_CERT, TLS_CACERT, PORT and CREDS_DIR to be exported
# Update the config to add keys etc to the config
${RWO_BASE_PATH}/gluster/updateconf ./plugin/config.json
# move the updated config to plugin
mv config.json ./plugin/config.json

echo -e "==> Remove existing plugin : ${PLUGIN_IMAGE} if exists"
docker plugin rm -f ${PLUGIN_IMAGE} || true
echo -e "==> Create new plugin : ${PLUGIN_IMAGE} from ./plugin"
docker plugin create ${PLUGIN_IMAGE} ./plugin

echo -e "==> Enable plugin ${PLUGIN_IMAGE}"
docker plugin enable ${PLUGIN_IMAGE}
