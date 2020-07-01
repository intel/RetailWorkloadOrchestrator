#!/bin/bash

docker run -d --privileged --name rwo_rngd edge/rngd:1.0 # Needed to create entropy more quickly
export UUID=$(docker run -i --rm --entrypoint="" edge/console-alpine:1.0 uuidgen)
if grep -q localhost /etc/hosts; then echo "" > /dev/null; else echo "127.0.0.1       localhost" >> /etc/hosts; fi
/usr/local/bin/docker-compose -p rwo -f /opt/rwo/compose/docker-compose.yml up -d
