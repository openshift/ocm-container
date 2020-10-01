#FROM ocm-container
FROM fedora:latest

ENV I_AM_IN_CONTAINER="I-am-in-container"

RUN yum -y install \
    bash-completion \
    findutils \
    fzf \
    git \
    golang \
    jq \
    make \
    procps-ng \
    rsync \
    sshuttle \
    vim-enhanced \
    wget;

RUN yum clean all;

ADD ./container-setup /container-setup

WORKDIR /container-setup

ARG osv4client=openshift-client-linux-4.3.12.tar.gz
ARG moactlversion=v0.0.5
ARG awsclient=awscli-exe-linux-x86_64.zip
ARG osdctlversion=v0.2.0
ARG veleroversion=v1.5.1

RUN ./install.sh

