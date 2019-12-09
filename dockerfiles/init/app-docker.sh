#!/bin/sh

echo "Trying to run dockerd"
[ -f /var/run/docker.pid ] &&  echo "Remove Stale pid file"  && rm -f /var/run/docker.pid
[ -f /var/run/docker.sock ] && echo "Remove Stale socket"  && rm -f /var/run/docker.sock

/usr/local/bin/dockerd-entrypoint.sh $@
