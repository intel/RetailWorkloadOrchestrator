#!/bin/sh
if [ -z "${SERF_LOG_LEVEL}" ]; then
    SERF_LOG_LEVEL=info
fi
cat << EOF > ${SERF_CONFIG_DIR}/010-logging.json
{
    "log_level" : "${SERF_LOG_LEVEL}"
}
EOF
# serf/docker-entrypoint.d/010-logging.sh
