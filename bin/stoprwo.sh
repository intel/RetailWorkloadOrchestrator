#!/bin/bash

/usr/local/bin/docker-compose -p rwo -f /opt/rwo/compose/docker-compose.yml down -v

for x in $(ls /mnt/);
do echo $x && umount /mnt/$x
done
