FROM centos:7
  
ARG GLUSTER_VER=7
ARG DUMBINIT_VER=1.2.2
ARG GLUSTERFS_SERVER_VERSION=7.1-1.el7.x86_64
ARG GLUSTER_GID=5000
ARG GLUSTER_UID=5000

RUN groupadd -g ${GLUSTER_GID} gluster \
    && adduser -g gluster -u ${GLUSTER_UID} gluster

RUN yum update -y -v \
        && yum install -y centos-release-gluster${GLUSTER_VER} \
        && yum install -y \
           glusterfs-server-${GLUSTERFS_SERVER_VERSION} \
        && yum clean all \
        && rm -rf /var/cache/yum \
        && rpm -e --nodeps tcp_wrappers-libs.x86_64

RUN curl -sSL -o /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v${DUMBINIT_VER}/dumb-init_${DUMBINIT_VER}_amd64 \
        && chmod +x /usr/local/bin/dumb-init

RUN rm -fr /var/lib/glusterd/ && mkdir -p /var/lib/glusterd/

WORKDIR /

ENTRYPOINT ["dumb-init", "gluster"]
CMD ["help"]
