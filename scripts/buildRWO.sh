#!/bin/bash

set -u

cd dockerfiles/

run "(1/7) Building Docker Image RWO Alpine Console" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/console-alpine:1.0 -f ./Dockerfile.console-alpine ." \
	${LOG_FILE}

run "(2/7) Building Docker Image GlusterFS" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs:5 -f ./Dockerfile.glusterfs-server ." \
	${LOG_FILE}

run "(3/7) Building Docker Image GlusterFS REST" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-rest:5 -f ./Dockerfile.glusterfs-rest ../" \
	${LOG_FILE}

run "(4/7) Building Docker Image Serf" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/serf:0.8.4 -f ./Dockerfile.serf ../" \
	${LOG_FILE}

run "(5/7) Building Docker Image Dynamic Hardware Orchestrator" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/dho:1.0 -f ./Dockerfile.dho ../" \
	${LOG_FILE}

run "(6/7) Building Docker App Docker" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/app-docker:1.0 -f ./Dockerfile.appdocker ." \
	${LOG_FILE}

cd - > /dev/null

cd glusterfs-plugin/

run "(7/7) Building Docker Image Gluster Plugin" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-plugin:1.0  -f ./Dockerfile ." \
	${LOG_FILE}

cd - > /dev/null


# preload images to app-docker
dockerimagelist="edge/glusterfs-plugin:1.0 edge/console-alpine:1.0"

mkdir -p var/lib/app-docker && \

docker run -d --privileged --name app-docker-build  -v /var/lib/app-docker:/var/lib/docker edge/app-docker:1.0  dockerd

## add images to app-docker
for image in $dockerimagelist; do
	 docker save $image | docker exec -i app-docker-build docker load
done

docker rm -f app-docker-build

# save a snapshot of app-docker
rsync -a /var/lib/app-docker /opt/rwo/
