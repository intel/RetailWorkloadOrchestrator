#!/bin/sh
umount ${GLUSTER_MOUNT_PATH}/${GLUSTER_VOLUME_NAME}
if [ -d ${GLUSTER_MOUNT_PATH}/${GLUSTER_VOLUME_NAME} ] && (! mountpoint ${GLUSTER_MOUNT_PATH}/${GLUSTER_VOLUME_NAME}); then
	# GLUSTER_CLUSTER_ADDR=$(ip -o -4 addr list ${GLUSTER_CLUSTER_IFACE} | head -1 | awk '{print $4}' | cut -d/ -f1)
	mount -t glusterfs ${GLUSTER_CLUSTER_ADDR}:/${GLUSTER_VOLUME_NAME} ${GLUSTER_MOUNT_PATH}/${GLUSTER_VOLUME_NAME}
fi
# gluster/server-entrypoint.sh
