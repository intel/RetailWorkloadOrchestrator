#!/usr/bin/dumb-init /bin/sh
set -e
SERF_CONFIG_DIR=${SERF_CONFIG_DIR?required}
SERF_DATA_DIR=${SERF_DATA_DIR?required}
SERF_VOLUME_CONFIG_PATH=${SERF_VOLUME_CONFIG_PATH?required}
SERF_HANDLERS_PATH=${SERF_HANDLERS_PATH?required}
SERF_CONFIG_FILE=${SERF_CONFIG_FILE?required}
SERF_CONFIG_FILE_PATH=${SERF_CONFIG_DIR}/${SERF_CONFIG_FILE}

# check if config file doesn't exist locally
if [ ! -f $SERF_CONFIG_FILE_PATH ]; then 
  # check if config file doesn't exist in usb-mount volume
  if [ -f ${SERF_VOLUME_CONFIG_PATH}/${SERF_CONFIG_FILE} ]; then
      cp ${SERF_VOLUME_CONFIG_PATH}/${SERF_CONFIG_FILE} ${SERF_CONFIG_DIR}
  else    
    # conf was not copied over from the usb-mount volume and doesn't exist locally, create it
    touch ${SERF_CONFIG_FILE_PATH}
    key="$(serf keygen)"
    echo '{"discover":"'${HOSTNAME}'", "encrypt_key":"'${key}'", "event_handlers":["'${SERF_HANDLERS_PATH}'"] }' > ${SERF_CONFIG_FILE_PATH}  
  fi
fi

if [ "$1" = 'agent' ]; then
    if [ -d "${0%.*}.d" ]; then
        for f in $(find "${0%.*}.d" -type f | sort); do
            case "$f" in
                *.sh)   echo "# $0: source $f" >&2; . "$f";;
                *)      echo "# $0: ignore $f" >&2;;
            esac
        done
    fi
    shift
    set -- agent "-config-dir=$SERF_CONFIG_DIR" "$@"
fi
echo "$@"
exec serf "$@"
# serf/docker-entrypoint.sh