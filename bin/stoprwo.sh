#!/bin/bash

docker-compose -p rwo -f /opt/rwo/compose/docker-compose.yml down -v
docker stop rwo_rngd && docker rm rwo_rngd

for x in $(ls /mnt/); do
    umount /mnt/$x >/dev/null
done
