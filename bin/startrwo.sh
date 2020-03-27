#!/bin/bash

export UUID=$(uuidgen)
if grep -q localhost /etc/hosts; then echo "" > /dev/null; else echo "127.0.0.1       localhost" >> /etc/hosts; fi
/opt/rwo/bin/reset > /dev/null
/usr/local/bin/docker-compose -p rwo -f /opt/rwo/compose/docker-compose.yml up -d
