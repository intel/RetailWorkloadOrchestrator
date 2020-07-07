#!/bin/sh
if [ ! -z "${SERF_BIND}" ]; then
    SERF_BIND_HOST=${SERF_BIND%:*}
    SERF_BIND_PORT=${SERF_BIND##*:}
    test "${SERF_BIND}" = "${SERF_BIND_HOST}" && unset SERF_BIND_PORT
fi
if [ -z "${SERF_BIND_HOST}" ]; then
    SERF_BIND_HOST=0.0.0.0
fi
if [ -z "${SERF_BIND_PORT}" ]; then
    SERF_BIND_PORT=7946
fi
if [ ! -z "${SERF_RPC_ADDR}" ]; then
    SERF_RPC_HOST=${SERF_RPC_ADDR%:*}
    SERF_RPC_PORT=${SERF_RPC_ADDR##*:}
    test "${SERF_RPC_ADDR}" = "${SERF_RPC_HOST}" && unset SERF_RPC_PORT
fi
if [ -z "${SERF_RPC_HOST}" ]; then
    SERF_RPC_HOST=127.0.0.1
fi
if [ -z "${SERF_RPC_PORT}" ]; then
    SERF_RPC_PORT=7373
fi
if [ ! -z "${SERF_ADVERTISE}" ]; then
    SERF_ADVERTISE_HOST=${SERF_ADVERTISE%:*}
    SERF_ADVERTISE_PORT=${SERF_ADVERTISE##*:}
    test "${SERF_ADVERTISE}" = "${SERF_ADVERTISE_HOST}" && unset SERF_ADVERTISE_PORT
fi
if [ -z "${SERF_ADVERTISE_HOST}" ] || [ "${SERF_ADVERTISE_HOST}" = "0.0.0.0" ]; then
    if [ -z "${SERF_ADVERTISE_IFACE}" ]; then
        # SERF_ADVERTISE_IFACE=$(ip -o -4 addr list | grep -v docker | grep global | awk '{print $2}'| head -n1)
        SERF_ADVERTISE_IFACE=$(ip route show 0.0.0.0/0 | awk '{print $5}')
    fi
    for LAN in ${SERF_ADVERTISE_IFACE}; do
        SERF_ADVERTISE_HOST=$(ip -o -4 addr list ${LAN} | head -1 | awk '{print $4}' | cut -d/ -f1)
        break
    done
else
    SERF_ADVERTISE_IFACE=$(ip -o -4 addr list | grep ${SERF_ADVERTISE_HOST} | head -1 | awk '{print $2}')
fi
if [ -z "${SERF_ADVERTISE_PORT}" ]; then
    SERF_ADVERTISE_PORT=${SERF_BIND_PORT}
fi
export SERF_ADVERTISE_IFACE SERF_ADVERTISE_HOST SERF_ADVERTISE_PORT
export SERF_ADVERTISE="${SERF_ADVERTISE_HOST}:${SERF_ADVERTISE_PORT}"
export SERF_BIND="${SERF_BIND_HOST}:${SERF_BIND_PORT}"
export SERF_RPC_ADDR="${SERF_RPC_HOST}:${SERF_RPC_PORT}"
cat << EOF > ${SERF_CONFIG_DIR}/010-listen.json
{
    "advertise" : "${SERF_ADVERTISE}",
    "bind" : "${SERF_BIND}",
    "rpc_addr" : "${SERF_RPC_ADDR}"
}
EOF
# serf/docker-entrypoint.d/010-listen.sh
