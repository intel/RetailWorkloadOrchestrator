FROM golang AS build-env

COPY . /go/src/app

WORKDIR /go/src/app

RUN make build

FROM ubuntu:18.04
LABEL maintainer="Bryan J Rodriguez <bryan.j.rodriguez@intel.com>"

RUN apt-get update \
 && apt-get install -y software-properties-common \
 && add-apt-repository ppa:gluster/glusterfs-7 \
 && apt-get update \
 && apt-get install -y glusterfs-client \
 && mkdir -p /var/lib/docker-volumes/gluster /etc/docker-volumes/gluster \
 && apt-get autoremove -y \
 && apt-get autoclean -y \
 && apt-get clean -y \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build-env /go/src/app/docker-volume-gluster /usr/bin/docker-volume-gluster
COPY updateconf /usr/local/bin/updateconf

RUN ln -s /bin/umount /usr/bin/umount
RUN /usr/bin/docker-volume-gluster version

ENTRYPOINT [ "/usr/bin/docker-volume-gluster" ]
CMD [ "daemon" ]
