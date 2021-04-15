#FROM ocm-container
FROM fedora:latest

ENV I_AM_IN_CONTAINER="I-am-in-container"

RUN yum --assumeyes install \
    bash-completion \
    findutils \
    fzf \
    git \
    golang \
    jq \
    make \
    openssl \
    procps-ng \
    python-pip \
    python3-requests-kerberos \
    rsync \
    sshuttle \
    tmux \
    vim-enhanced \
    wget \
    && yum clean all;

ADD ./container-setup/install /container-setup/install

WORKDIR /container-setup/install

ARG awsclient=awscli-exe-linux-x86_64.zip
ARG osdctlversion=v0.4.0
ARG osv4client=openshift-client-linux-4.7.2.tar.gz
ARG rosaversion=v1.0.1
ARG veleroversion=v1.5.3

RUN ./install-aws.sh
RUN ./install-cluster-login.sh
RUN ./install-kube_ps1.sh
RUN ./install-oc.sh
RUN ./install-ocm.sh
RUN ./install-osdctl.sh
RUN ./install-rosa.sh
RUN ./install-velero.sh

ADD ./container-setup/utils /container-setup/utils
WORKDIR /container-setup/utils
RUN ./install-utils.sh

ENV PATH "$PATH:/root/utils"
RUN rm -rf /container-setup

WORKDIR /root
ENTRYPOINT ["/bin/bash"]
