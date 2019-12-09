#!/bin/sh
set -e

GLUSTERFS_REST_PORT=${GLUSTERFS_REST_PORT?required}
IP_ADDRESS=${GLUSTERFS_REST_IP?required}

user=`sed '1q;d' ${CREDS_DIR}/creds.txt`
password=`sed '2q;d' ${CREDS_DIR}/creds.txt`

# Check if you are on the gluster
PeerResponse=$(curl --insecure -X GET https://$user:$password@$IP_ADDRESS:${GLUSTERFS_REST_PORT}/api/1.0/peers)

# Parse the result for the client name
# EDGE_IFACE=$(ip -o -4 addr list | grep -v docker | grep global | awk '{print $2}'| head -n1)
EDGE_IFACE=$(ip route show 0.0.0.0/0 | awk '{print $5}')
EDGE_ADDR=$(ip -o -4 addr list ${EDGE_IFACE} | head -1 | awk '{print $4}' | cut -d/ -f1)
# If we don't see the client name in the list then we are not connected and need to connect using the peer endpoint
if [ -n "`echo $PeerResponse | grep -e $EDGE_ADDR`" ]; then
  echo "Already connected to the gluster server"
else
  echo "connecting to gluster pool"
  # if not connected then connect
  connectResponse=$(curl --insecure -X POST https://$user:$password@$IP_ADDRESS:${GLUSTERFS_REST_PORT}/api/1.0/peer/${EDGE_ADDR})
  echo $connectResponse
fi
