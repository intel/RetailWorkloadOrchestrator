#!/bin/bash
# This fiile should be run from RWO root folder or scripts folder
# Eg:
# 	bash scripts/testSerfHandlers.sh
#   				or
#   cd scripts && ./testSerfHandlers.sh
#
# For Debugging and testing we may not like to remove the downloaded packages everytime we run the script,
#
#       For the first time, do this:
#  	    - bash scripts/testSerfHandlers.sh debug
#       The above step will ensure that pakcages downloaded to the temporary path are not deleted.
#       Then subsequently, we can use the below command, to resuse the already downloaded packages
#           - bash scripts/testSerfHandlers.sh fast

red=`tput setaf 1`
green=`tput setaf 2`
blue=`tput setaf 4`
C_RED='\e[31m'
C_GREEN='\e[32m'
T_RESET='\e[0m'
T_BOLD='\e[1m'

PWD=`pwd`
echo "Build started $PWD"

T_ERR_ICON="[${T_BOLD}${C_RED}✗${T_RESET}]"
T_OK_ICON="[${T_BOLD}${C_GREEN}✓${T_RESET}]"

# Update serf path
if  [[ -e .git ]]; then
  SERF_PATH=${PWD}/serf/handlers
else
# if run from inside scripts folder
  SERF_PATH=${PWD}/../serf/handlers
fi

build_status_accumulated=0
lint_status_accumulated=0

unit_test() {
    file=$1
    # Check $i is a file
    file_in_src=`echo $file | awk -F"src/" '{ print $2}'`
    test_file=`echo $file_in_src| sed 's/.go/_test.go/g'`
    echo $file_in_src" "$test_file

    if [ -f ${file} ]; then
        OUTPUT=`echo ${file_in_src} | grep  "\.go" | awk -F"/" '{ print $NF}' | sed 's/.go//g'`
        mkdir -p ${SERF_PATH}/bin

        HELPERS_MOCK_PATH=${SERF_PATH}/tests/mock_packages/helpers
        RWOGLUSTER_MOCK_PATH=${SERF_PATH}/tests/mock_packages/rwogluster

        docker run --rm  -e "CGO_ENABLED=0" -v ${SERF_PATH}/.gopath/src/golang.org:/data/src/golang.org -v ${SERF_PATH}/.gopath/src/github.com:/data/src/github.com  -e "GOPATH=/data/" -v ${RWOGLUSTER_MOCK_PATH}:/data/src/rwogluster -v ${HELPERS_MOCK_PATH}:/data/src/helpers  -v ${SERF_PATH}/src/member-update-x:/data/src/memberupdatex  -v ${SERF_PATH}/bin:/data/bin   -v ${SERF_PATH}/src:/data/serf golang:latest go test -v /data/serf/${file_in_src} /data/serf/${test_file}

    fi
}

cleanup_temp_gopath() {
    echo "Removing temporary go path."
    # Remove temporary gopath
    docker run --net=host	 --rm -e "GOPATH=/data/" -e "http_proxy=$http_proxy" -e "https_proxy=$http_proxy" -v ${SERF_PATH}/.gopath:/data/ alpine:3.9 sh -c "cd /data && rm -rf *"
    rmdir ${SERF_PATH}/.gopath
}

if [ ! -z $http_proxy ]; then
        PROXY_ARGS="--env \"http_proxy=$http_proxy\" -e \"https_proxy=$http_proxy\""
else
        PROXY_ARGS=""
fi

# Make temporary gopath
# later clean that up

if [[ -e ${SERF_PATH}/.gopath ]]; then
    if [[ $1 = "reset" ]]; then
        cleanup_temp_gopath
        exit
    else
        echo "Packages exist. Not pulling"
    fi
else

    mkdir -p ${SERF_PATH}/.gopath
    echo -e  "Downloading Packages.... ${T_RESET} "
    echo -e  "Do not abort ${T_RESET} "
    docker run --net=host	 --rm -e "GOPATH=/data/" ${PROXY_ARGS} -v ${SERF_PATH}/.gopath:/data/ golang:latest go get golang.org/x/sys/unix
    #echo -e  "${blue} GO ${T_OK_ICON} ${T_RESET} "
    docker run --net=host	 --rm -e "GOPATH=/data/" ${PROXY_ARGS} -v ${SERF_PATH}/.gopath:/data/ golang:latest go get github.com/sirupsen/logrus
    #echo -e  "${blue} LOGRUS ${T_OK_ICON} ${T_RESET}"
    docker run --net=host	 --rm -e "GOPATH=/data/" ${PROXY_ARGS} -v ${SERF_PATH}/.gopath:/data/ golang:latest go get github.com/hashicorp/serf/client
    #echo -e  "${blue} SERF CLIENT ${T_OK_ICON} ${T_RESET}"
    docker run --net=host	 --rm -e "GOPATH=/data/" ${PROXY_ARGS} -v ${SERF_PATH}/.gopath:/data/ golang:latest go get github.com/docker/docker
    # echo -e  "${blue} DOCKER CLIENT ${T_OK_ICON} ${T_RESET}"
fi

echo -e  "Begin Unittest ${T_RESET} "

unit_test ${SERF_PATH}/src/memberjoin.go
unit_test ${SERF_PATH}/src/memberupdate.go
unit_test ${SERF_PATH}/src/memberfailed.go
unit_test ${SERF_PATH}/src/membercleanup.go

if [[ $1 = "debug" ]]; then
    echo "Go path is not removed."
else
    cleanup_temp_gopath
fi;

