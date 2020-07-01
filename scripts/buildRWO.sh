#!/bin/bash

set -u

cd dockerfiles/

run "Pulling Base Docker Images" \
	"docker pull docker:19.03.0 && docker pull docker:19.03.0-dind" \ 
	${LOG_FILE} # Needed for the Arbiter

run "(1/7) Building Docker Image RWO Alpine Console" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/console-alpine:1.0 -f ./Dockerfile.console-alpine ." \
	${LOG_FILE}

run "(2/7) Building Docker Image GlusterFS" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs:7 -f ./Dockerfile.glusterfs-server ." \
	${LOG_FILE}
	
run "(3/7) Building Docker Image GlusterFS REST" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-rest:7 -f ./Dockerfile.glusterfs-rest ../" \
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
	"cp ../gluster/updateconf . && docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-plugin:1.0  -f ./Dockerfile ." \
	${LOG_FILE}

cd - > /dev/null
