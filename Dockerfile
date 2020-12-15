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
    python-pip \
    python3-requests-kerberos \
    rsync \
    sshuttle \
    vim-enhanced \
    wget \
    && yum clean all;

ADD ./container-setup/install /container-setup/install

WORKDIR /container-setup/install

ARG osv4client=openshift-client-linux-4.3.12.tar.gz
ARG rosaversion=v0.0.16
ARG awsclient=awscli-exe-linux-x86_64.zip
ARG osdctlversion=v0.2.0
ARG veleroversion=v1.5.1

RUN ./install-rosa.sh
RUN ./install-ocm.sh
RUN ./install-oc.sh
RUN ./install-aws.sh
RUN ./install-kube_ps1.sh
RUN ./install-osdctl.sh
RUN ./install-velero.sh
RUN ./install-cluster-login.sh

ADD ./container-setup/utils /container-setup/utils
WORKDIR /container-setup/utils
RUN ./install-utils.sh

ENV PATH "$PATH:/root/utils"
RUN rm -rf /container-setup

WORKDIR /root
ENTRYPOINT ["/bin/bash"]
