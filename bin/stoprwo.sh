#!/bin/bash

docker-compose -p rwo -f /opt/rwo/compose/docker-compose.yml down -v

for x in $(ls /mnt/); do
    umount /mnt/$x >/dev/null
done
