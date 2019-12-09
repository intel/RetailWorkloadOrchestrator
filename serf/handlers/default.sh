#!/bin/sh
HANDLER_DIR=$(dirname $0)
HANDLER_PATH=${HANDLER_DIR}/${SERF_EVENT}
if [ ! -z "${SERF_QUERY_NAME}" ]; then
    HANDLER_PATH=${HANDLER_PATH}/${SERF_QUERY_NAME}
elif [ ! -z "${SERF_USER_EVENT}" ]; then
    HANDLER_PATH=${HANDLER_PATH}/${SERF_USER_EVENT}
fi

touch /var/log/serf.log
chmod 644 /var/log/serf.log

if [ -z "${SERF_QUERY_NAME}" ]; then
printf '\n%-16s %-10s init=%-20s role=%-10s\n\tgluster=%s\n\tswarm=%s\n' \
    "${SERF_EVENT}" "${SERF_SELF_NAME}" "${SERF_TAG_INIT}" "${SERF_TAG_ROLE}" \
    "${SERF_TAG_GLUSTER}" "${SERF_TAG_SWARM}" < /dev/null 2>&1 | tee -a ${SERF_TTY} /var/log/serf.log
fi
if [ -f ${HANDLER_PATH} ] && [ -x ${HANDLER_PATH} ]; then
    exec ${HANDLER_PATH} 2>&1 | tee -a /var/log/serf.log
fi
# serf/handlers/default.sh
