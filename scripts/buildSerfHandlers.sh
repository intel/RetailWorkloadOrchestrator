#!/bin/bash
# This fiile should be run from RWO root folder or scripts folder
# Eg:
# 	bash scripts/buildSerfHandlers.sh
#   				or
#   cd scripts && ./buildSerfHandlers.sh
#
# During development or debugging, we don't want to pull packages again and again
# so, we can use this command to keep the packages persistent
#   bash scripts/buildSerfHandlers.sh debug

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

compile_updateconf() {
        echo "Compile updateconf.go"
        docker run --rm -e "GOPATH=/data/" -v ${SERF_PATH}/../../gluster:/data/src cytopia/golint -set_exit_status /data/src/updateconf.go
        docker run --rm -e "CGO_ENABLED=0" -v ${SERF_PATH}/../../gluster:/data/src golang:latest go build -a -installsuffix cgo -o /data/src/updateconf /data/src/updateconf.go
}

lint_handlers() {
	file=$1
    # Check $i is a file
    file_in_src=`echo $file | awk -F"src/" '{ print $2}'`

    if [ -f ${file} ]; then


        docker run --rm -e "GOPATH=/data/" -v ${SERF_PATH}:/data cytopia/golint -set_exit_status /data/src/${file_in_src}
        lint_status=$?

        if [ $lint_status -gt 0 ]; then
            echo -e " Lint ${T_ERR_ICON} ${T_RESET}"
        else
            echo -e " Lint ${T_OK_ICON} ${T_RESET}"
        fi

        lint_status_accumulated=`expr $lint_status + $lint_status_accumulated`
    fi
}

compile_handlers() {
	file=$1
    # Check $i is a file
    file_in_src=`echo $file | awk -F"src/" '{ print $2}'`

    if [ -f ${file} ]; then
        OUTPUT=`echo ${file_in_src} | grep  "\.go" | awk -F"/" '{ print $NF}' | sed 's/.go//g'`
        mkdir -p ${SERF_PATH}/bin

#        docker run --rm -e "GOPATH=/data/" -v ${SERF_PATH}:/data golang:latest go build -o /data/bin/${OUTPUT} /data/src/${file_in_src}

        docker run --rm -e "CGO_ENABLED=0" -v ${SERF_PATH}/.gopath/src/golang.org:/data/src/golang.org  -v ${SERF_PATH}/.gopath/src/github.com:/data/src/github.com -e "GOPATH=/data/" -v ${SERF_PATH}/../../glusterfs-lib/rwogluster:/data/src/rwogluster   -v ${SERF_PATH}/src/helpers:/data/src/helpers -v ${SERF_PATH}/src/member-update-x:/data/src/memberupdatex -v ${SERF_PATH}/bin:/data/bin -v ${SERF_PATH}/src:/data/serf golang:latest go build -a -installsuffix cgo -o /data/bin/${OUTPUT} /data/serf/${file_in_src}

        build_status=$?
        if [ $build_status -gt 0 ]; then
            echo -e " Build ${T_ERR_ICON} ${T_RESET}"
        else
            echo -e " Build ${T_OK_ICON} ${T_RESET}"
        fi

        build_status_accumulated=`expr $build_status + $build_status_accumulated`
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
echo -e  "Begin Compilation ${T_RESET} "
fi

compile_updateconf

# compile the handler source files
for i in `ls ${SERF_PATH}/src/*.go | grep -v test`
do
    echo "$i"  | awk -F"zeroConf/" '{ print $2}'
    lint_handlers $i
    compile_handlers $i
done

# compile query source files
# compile the handler source files
for i in `ls ${SERF_PATH}/src/query/*.go`
do
    echo "$i" | awk -F"zeroConf/" '{ print $2}'
    lint_handlers $i
    compile_handlers $i
done

# lint helpers source files
for i in `ls ${SERF_PATH}/src/helpers/*.go`
do
    lint_handlers $i
done

# lint member-update-x source files
for i in `ls ${SERF_PATH}/src/member-update-x/*.go`
do
    lint_handlers $i
done

if [[ $1 = "debug" ]]; then
    echo "Go path is not removed."
else
    cleanup_temp_gopath
fi;

## Check for linting problem and report error

if [ $build_status_accumulated -gt 0 ] || [ $lint_status_accumulated -gt 0 ]; then
    echo "${red} Lint Status "${lint_status_accumulated}
    echo "${red} Build Status "${build_status_accumulated}

	exit 1
fi
## Check for build problem
