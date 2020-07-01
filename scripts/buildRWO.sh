#!/bin/bash

set -u

cd dockerfiles/

run "(1/8) Building Docker Image RWO Alpine Console" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/console-alpine:1.0 -f ./Dockerfile.console-alpine ." \
	${LOG_FILE}

run "(2/8) Building Docker Image GlusterFS" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs:7 -f ./Dockerfile.glusterfs-server ." \
	${LOG_FILE}
	
run "(3/8) Building Docker Image GlusterFS REST" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-rest:7 -f ./Dockerfile.glusterfs-rest ../" \
	${LOG_FILE}
	
run "(4/8) Building Docker Image Serf" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/serf:0.8.4 -f ./Dockerfile.serf ../" \
	${LOG_FILE}

run "(5/8) Building Docker Image Dynamic Hardware Orchestrator" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/dho:1.0 -f ./Dockerfile.dho ../" \
	${LOG_FILE}

run "(6/8) Building Docker Image App Docker" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/app-docker:1.0 -f ./Dockerfile.appdocker ." \
	${LOG_FILE}

run "(7/8) Building Docker Image Random Number Generator" \
	"docker build --rm ${DOCKER_BUILD_ARGS} -t edge/rngd:1.0 -f ./Dockerfile.rngd ." \
	${LOG_FILE}

cd - > /dev/null

cd glusterfs-plugin/

run "(8/8) Building Docker Image Gluster Plugin" \
	"cp ../gluster/updateconf . && docker build --rm ${DOCKER_BUILD_ARGS} -t edge/glusterfs-plugin:1.0  -f ./Dockerfile ." \
	${LOG_FILE}

cd - > /dev/null
