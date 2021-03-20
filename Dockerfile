#FROM ocm-container
FROM fedora:latest

ENV I_AM_IN_CONTAINER="I-am-in-container"

# the deletion of 'vim-minimal' is required for package vim-enhanced-2:8.2.2143-1
# see https://unix.stackexchange.com/a/120226
RUN yum --assumeyes remove \
    vim-minimal \ 
    && yum --assumeyes install \
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
    tmux \
    vim-enhanced \
    wget \
    && yum clean all;

ADD ./container-setup/install /container-setup/install

WORKDIR /container-setup/install

ARG awsclient=awscli-exe-linux-x86_64.zip
ARG osdctlversion=v0.4.0
ARG osv4client=openshift-client-linux-4.3.12.tar.gz
ARG rosaversion=v0.0.16
ARG veleroversion=v1.5.1

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
