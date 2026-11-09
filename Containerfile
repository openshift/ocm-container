# Red Hat Hardened Image (Project Hummingbird) bases, using floating tags so Red Hat manages updates.
# The runtime stages use core-runtime; the build/installer stages use the -builder variant (dnf, rpm,
# curl), which is not part of the shipped image.
ARG RUNTIME_BASE_IMAGE=registry.access.redhat.com/hi/core-runtime:latest
ARG INSTALLER_BASE_IMAGE=registry.access.redhat.com/hi/core-runtime:latest-builder
FROM ${INSTALLER_BASE_IMAGE} as tools-base
USER 0
ENV HOME=/root
ARG OUTPUT_DIR="/opt"

RUN dnf --assumeyes install gzip jq tar python3.14 python3-pip
RUN python3.14 -m pip install --no-cache-dir requests

# Adds Platform Conversion Tool for arm64/x86_64 compatibility
# need to add this a second time to add it to the builder image
COPY utils/dockerfile_assets/platforms.sh /usr/local/bin/platform_convert
COPY utils/dockerfile_assets/github_dl.py /usr/local/bin/github_dl

### BACKPLANE TOOLS - download SRE standad binaries to a temporary container
FROM tools-base as backplane-tools
ARG OUTPUT_DIR="/opt"

ARG BACKPLANE_TOOLS_VERSION="tags/v1.6.1"
ENV BACKPLANE_TOOLS_URL_SLUG="openshift/backplane-tools"
ENV BACKPLANE_TOOLS_URL="https://api.github.com/repos/${BACKPLANE_TOOLS_URL_SLUG}/releases/${BACKPLANE_TOOLS_VERSION}"
ENV BACKPLANE_TOOLS_CHECKSUM_FILE="checksums.txt"
ENV BACKPLANE_TOOLS_CHECKSUM_ALGORITHM="sha256"
ENV BACKPLANE_TOOLS_PLATFORM_PREFIX="linux_"
ENV BACKPLANE_BIN_DIR="/root/.local/bin/backplane"

RUN mkdir -p /backplane-tools
WORKDIR /backplane-tools

# Download the checksum and binary, and validate them
RUN --mount=type=secret,id=GITHUB_TOKEN \
    --mount=type=secret,id=read-only-github-pat/token \
    github_dl download --url ${BACKPLANE_TOOLS_URL} --checksum_file ${BACKPLANE_TOOLS_CHECKSUM_FILE} --checksum_algorithm ${BACKPLANE_TOOLS_CHECKSUM_ALGORITHM} --platform ${BACKPLANE_TOOLS_PLATFORM_PREFIX}$(platform_convert "@@PLATFORM@@" --amd64 --arm64)

# Extract the binary tarball
RUN tar --extract --gunzip --no-same-owner --directory "/usr/local/bin"  --file *.tar.gz

# Install only the backplane-tools packages ocm-container actually ships.
# (See ROSAENG-70443: `install all` pulled in every tool backplane-tools
# manages, including ones this image never copies out of this stage, which
# meant an unrelated tool's upstream release-asset rename could break the
# whole build. `omc`, `ocm-addons`, `rosa` and `butane` are evaluated and
# intentionally dropped here -- see the PR/ticket for the keep/drop rationale
# per tool. The `backplane-tools` binary itself is still copied into the
# final image below so any dropped package can be installed manually.)
ARG BACKPLANE_TOOLS_PACKAGES="aws oc ocm osdctl backplane-cli yq gcloud"
RUN --mount=type=secret,id=GITHUB_TOKEN \
    --mount=type=secret,id=read-only-github-pat/token \
    if [[ -f /run/secrets/read-only-github-pat/token ]]; then \
        echo "PAT FOUND: $(wc -c < /run/secrets/read-only-github-pat/token) bytes, first4=$(head -c 4 /run/secrets/read-only-github-pat/token)"; \
        GITHUB_TOKEN=$(cat /run/secrets/read-only-github-pat/token) /usr/local/bin/backplane-tools install ${BACKPLANE_TOOLS_PACKAGES}; \
    elif [[ -f /run/secrets/GITHUB_TOKEN ]]; then \
        echo "GITHUB_TOKEN FOUND: $(wc -c < /run/secrets/GITHUB_TOKEN) bytes, first4=$(head -c 4 /run/secrets/GITHUB_TOKEN)"; \
        GITHUB_TOKEN=$(cat /run/secrets/GITHUB_TOKEN) /usr/local/bin/backplane-tools install ${BACKPLANE_TOOLS_PACKAGES}; \
    else echo "nope" && /usr/local/bin/backplane-tools install ${BACKPLANE_TOOLS_PACKAGES} ;\
    fi

# Copy symlink sources from ./local/bin to $OUTPUT_DIR
RUN cp -Hv  ${BACKPLANE_BIN_DIR}/latest/* ${OUTPUT_DIR}

# copy aws cli assets
RUN cp -r ${BACKPLANE_BIN_DIR}/aws/*/aws-cli/dist /${OUTPUT_DIR}/aws_dist

# copy gcloud sdk assets (gcloud ships as a full SDK directory tree, not a
# single binary, so it needs the same dedicated-copy treatment as aws-cli
# above rather than the flattened copy done for single-binary tools)
RUN cp -r ${BACKPLANE_BIN_DIR}/gcloud/*/google-cloud-sdk /${OUTPUT_DIR}/gcloud_sdk

### Builder - Get or Build Individual Binaries
FROM tools-base as builder
ARG OUTPUT_DIR="/opt"

# jq is a pre-req for making parsing of download urls easier
RUN dnf --assumeyes --nodocs install \
      gcc \
      git \
      jq \
      make \
      tar \
      unzip

# Directory for the extracted binaries, etc; used in child images
RUN mkdir -p /${OUTPUT_DIR}

### Hummingbird runtime base
FROM ${RUNTIME_BASE_IMAGE} as runtime-base

### dnf-bootstrap: install dnf (only) into the runtime root, ship only the delta
FROM ${INSTALLER_BASE_IMAGE} as dnf-bootstrap
USER 0
RUN dnf --assumeyes install rsync
COPY --from=runtime-base / /runtime/
RUN dnf --installroot=/runtime --use-host-config --nodocs --setopt=install_weak_deps=False install -y dnf \
    && dnf --installroot=/runtime --use-host-config clean all
RUN cp -a /etc/yum.repos.d/*.repo /runtime/etc/yum.repos.d/
RUN cp -a /etc/pki/rpm-gpg/* /runtime/etc/pki/rpm-gpg/
RUN printf '%s\n' \
      '[ubi-baseos]' 'name=ubi-baseos' \
      'baseurl=https://cdn-ubi.redhat.com/content/public/ubi/dist/ubi10/10/$basearch/baseos/os' \
      'enabled=0' 'gpgcheck=1' 'gpgkey=https://access.redhat.com/security/data/fd431d51.txt' \
      '[ubi-appstream]' 'name=ubi-appstream' \
      'baseurl=https://cdn-ubi.redhat.com/content/public/ubi/dist/ubi10/10/$basearch/appstream/os' \
      'enabled=0' 'gpgcheck=1' 'gpgkey=https://access.redhat.com/security/data/fd431d51.txt' \
      > /runtime/etc/yum.repos.d/ubi-extras.repo
RUN printf 'install_weak_deps=False\ntsflags=nodocs\n' >> /runtime/etc/dnf/dnf.conf
RUN --mount=type=bind,from=runtime-base,source=/,target=/base \
    mkdir -p /delta && rsync -a --compare-dest=/base/ /runtime/ /delta/

### py-wheels: build Python packages with compiled extensions; only wheels are shipped
FROM ${INSTALLER_BASE_IMAGE} as py-wheels
USER 0
ARG O_MUST_GATHER_VERSION=""
RUN dnf --assumeyes --nodocs install python3.14 python3-pip python3.14-devel krb5-devel gcc \
    && mkdir /wheels \
    && python3.14 -m pip wheel --no-cache-dir --wheel-dir /wheels \
       o-must-gather${O_MUST_GATHER_VERSION} rh-aws-saml-login

### Pre-install stuff for final images
FROM runtime-base as base-update
# Hummingbird runs as a non-root user with HOME=/tmp; the toolbox runs as root
USER 0
ENV HOME=/root
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"

# Bootstrap dnf: only the delta (files added/changed by the dnf package) is shipped,
# not a whole-root overlay. This is the only use of --installroot in the build.
COPY --from=dnf-bootstrap /delta/ /
RUN dnf --version && dnf repolist

# Shell init for all three images: source /etc/bashrc (which loops over
# /etc/profile.d/*.sh, loading bash-completion once that package is installed)
# and put ~/.local/bin on PATH so sre-login and other utils are found.
RUN printf '%s\n' \
      '[ -f /etc/bashrc ] && . /etc/bashrc' \
      'export PATH="$HOME/.local/bin:$PATH"' \
      > /root/.bashrc \
    && printf '%s\n' \
      '[ -f ~/.bashrc ] && . ~/.bashrc' \
      > /root/.bash_profile

# Set an exposable port for the cluster console proxy
# Can be used with `-o "-P"` to map 9999 inside the container to a random port at runtime
ENV OCM_BACKPLANE_CONSOLE_PORT 9999
EXPOSE $OCM_BACKPLANE_CONSOLE_PORT
ENTRYPOINT ["/bin/bash"]

# Create a directory for the ocm config file
RUN mkdir -p /root/.config/ocm /etc/bash_completion.d

### Micro Image
FROM base-update as ocm-container-micro
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"
ARG BIN_DIR="/usr/local/bin"

# bash-completion is installed here (not in full) so that all three images
# get working tab completion: it supplies /etc/profile.d/bash_completion.sh
# which /etc/bashrc sources for interactive shells.
RUN dnf --assumeyes --nodocs install \
      bash-completion \
      bind-utils \
      jq \
    && dnf clean all

COPY --from=backplane-tools /${OUTPUT_DIR}/ocm           ${BIN_DIR}
RUN ocm completion > /etc/bash_completion.d/ocm

COPY --from=backplane-tools /${OUTPUT_DIR}/ocm-backplane ${BIN_DIR}
RUN ocm backplane completion bash > /etc/bash_completion.d/ocm-backplane

COPY --from=backplane-tools /${OUTPUT_DIR}/oc            ${BIN_DIR}
RUN oc completion bash > /etc/bash_completion.d/oc

ENV IO_OPENSHIFT_MANAGED_NAME="ocm-container"
ENV IO_OPENSHIFT_MANAGED_COMPONENT="micro"

### Final Minimal Image
FROM ocm-container-micro as ocm-container-minimal
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"
ARG BIN_DIR="/usr/local/bin"

COPY --from=backplane-tools /${OUTPUT_DIR}/aws_dist      /usr/local/aws-cli
RUN /usr/local/aws-cli/aws --version
RUN /usr/local/aws-cli/aws_completer bash > /etc/bash_completion.d/aws-cli

COPY --from=backplane-tools /${OUTPUT_DIR}/osdctl        ${BIN_DIR}
RUN osdctl completion bash --skip-version-check > /etc/bash_completion.d/osdctl

COPY --from=backplane-tools /${OUTPUT_DIR}/yq            ${BIN_DIR}
RUN yq --version

COPY --from=backplane-tools /${OUTPUT_DIR}/gcloud_sdk    /usr/local/gcloud-sdk
RUN ln -s /usr/local/gcloud-sdk/bin/gcloud ${BIN_DIR}/gcloud
RUN gcloud --version

# Ship the backplane-tools binary itself so any package dropped from
# BACKPLANE_TOOLS_PACKAGES above (eg. omc, ocm-addons, rosa, butane) can
# still be installed manually inside a running container on demand.
COPY --from=backplane-tools /usr/local/bin/backplane-tools ${BIN_DIR}
RUN backplane-tools --help > /dev/null

ENV IO_OPENSHIFT_MANAGED_NAME="ocm-container"
ENV IO_OPENSHIFT_MANAGED_COMPONENT="minimal"

# Binary builds for extra (full ocm-container) packages
FROM builder as omc-builder
ARG OUTPUT_DIR="/opt"
# Add `omc` utility to inspect must-gathers easily with 'oc' like commands
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG OMC_VERSION="tags/v3.17.0"
ENV OMC_URL_SLUG="gmeghnag/omc"
ENV OMC_URL="https://api.github.com/repos/${OMC_URL_SLUG}/releases/${OMC_VERSION}"
ENV OMC_CHECKSUM_FILE="checksums.txt"
ENV OMC_CHECKSUM_ALGORITHM="md5"
ENV OMC_PLATFORM_PREFIX="Linux_"
ENV OMC_PLATFORM_SUFFIX=".tar.gz"

# Install omc
RUN mkdir /omc
WORKDIR /omc

# Download the checksum and binary, and validate them
RUN --mount=type=secret,id=GITHUB_TOKEN \
    --mount=type=secret,id=read-only-github-pat/token \
    github_dl download --url ${OMC_URL} --checksum_file checksums.txt --checksum_algorithm ${OMC_CHECKSUM_ALGORITHM} --platform ${OMC_PLATFORM_PREFIX}$(platform_convert "@@PLATFORM@@" --x86_64 --arm64)${OMC_PLATFORM_SUFFIX}

# Extract the binary tarball
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} omc --file *.tar.gz
RUN chmod -R +x /${OUTPUT_DIR}

FROM builder as jira-builder
ARG OUTPUT_DIR="/opt"
# Add `jira` utility for working with OHSS tickets
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG JIRA_VERSION="tags/v1.7.0"
ENV JIRA_URL_SLUG="ankitpokhrel/jira-cli"
ENV JIRA_URL="https://api.github.com/repos/${JIRA_URL_SLUG}/releases/${JIRA_VERSION}"
ENV JIRA_CHECKSUM_FILE="checksums.txt"
ENV JIRA_CHECKSUM_ALGORITHM="sha256"
ENV JIRA_PLATFORM_PREFIX="linux_"

RUN mkdir /jira
WORKDIR /jira

# Download the checksum and binary, and validate them
RUN --mount=type=secret,id=GITHUB_TOKEN \
    --mount=type=secret,id=read-only-github-pat/token \
    github_dl download --url ${JIRA_URL} --checksum_file checksums.txt --checksum_algorithm ${JIRA_CHECKSUM_ALGORITHM} --platform ${JIRA_PLATFORM_PREFIX}$(platform_convert "@@PLATFORM@@" --x86_64 --arm64)

# Extract the binary tarball
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} --strip-components=2 */bin/jira --file *.tar.gz
RUN chmod -R +x /${OUTPUT_DIR}

FROM builder as oc-nodepp-builder
ARG OUTPUT_DIR="/opt"
# Add `oc-nodepp` utility
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG NODEPP_VERSION="tags/v0.1.2"
ENV NODEPP_URL_SLUG="mrbarge/oc-nodepp"
ENV NODEPP_URL="https://api.github.com/repos/${NODEPP_URL_SLUG}/releases/${NODEPP_VERSION}"
ENV NODEPP_CHECKSUM_FILE="checksums.txt"
ENV NODEPP_CHECKSUM_ALGORITHM="sha256"
ENV NODEPP_PLATFORM_PREFIX="Linux_"

# Install oc-nodepp
RUN mkdir /nodepp
WORKDIR /nodepp

# Download the checksum and binary, and validate them
RUN --mount=type=secret,id=GITHUB_TOKEN \
    --mount=type=secret,id=read-only-github-pat/token \
    github_dl download --url ${NODEPP_URL} --checksum_file ${NODEPP_CHECKSUM_FILE} --checksum_algorithm ${NODEPP_CHECKSUM_ALGORITHM} --platform ${NODEPP_PLATFORM_PREFIX}$(platform_convert "@@PLATFORM@@" --x86_64 --arm64)

# Extract the binary tarball
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} oc-nodepp --file *.tar.gz
RUN chmod +x /${OUTPUT_DIR}/oc-nodepp

###########################
## Build the final image ##
###########################
# This is based on the first image build, with the yum packages installed
FROM ocm-container-minimal as ocm-container
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"
ARG BIN_DIR="/usr/local/bin"

# Packages from the Hummingbird repos (ca-certificates, for update-ca-trust, replaces the base's
# ca-certificates-bundle, hence --allowerasing)
RUN dnf --assumeyes --nodocs install --allowerasing \
      ca-certificates \
      crun \
      curl \
      findutils \
      fuse-overlayfs \
      gawk \
      git \
      grep \
      gzip \
      krb5-workstation \
      make \
      openssl \
      podman \
      procps-ng \
      python3.14 \
      python3-pip \
      rsync \
      sed \
      tar \
      util-linux \
      xz \
    && dnf clean all

# socat, wget and vim are not in the Hummingbird repos: install the signed UBI 10 RPMs from the
# (otherwise disabled) UBI repos
RUN dnf --assumeyes --nodocs \
      --disablerepo='public-hummingbird*' --enablerepo='ubi-*' install \
      socat \
      vim-enhanced \
      vim-minimal \
      wget \
    && dnf clean all

# ssm plugin (Hummingbird's rpm rejects this package's missing header signature; UBI's accepted it)
# TODO: pin a version and verify a checksum instead of pulling /latest/ unchecked
COPY --from=tools-base /usr/local/bin/platform_convert ${BIN_DIR}
RUN rpm -i --nosignature $(platform_convert https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_@@PLATFORM@@/session-manager-plugin.rpm --arm64 --custom-amd64 64bit)
RUN rm ${BIN_DIR}/platform_convert

# Python packages from pre-built wheels (no compiler or -devel package ships)
RUN --mount=type=bind,from=py-wheels,source=/wheels,target=/tmp/wheels \
    python3.14 -m pip install --no-cache-dir --no-index --find-links /tmp/wheels \
      o-must-gather rh-aws-saml-login

### podman container config
# Overlay over overlay is often denied by the kernel, so this creates non overlay volumes to be used within the container.
VOLUME /var/lib/containers

# copy storage.conf to enable fuse-overlayfs storage.
COPY utils/dockerfile_assets/storage.conf /etc/containers/storage.conf

# add containers.conf file to make sure containers run easier.
COPY utils/dockerfile_assets/containers.conf /etc/containers/containers.conf

# Copy previously acquired binaries into the $PATH
WORKDIR /

COPY --from=jira-builder      /${OUTPUT_DIR}/jira      ${BIN_DIR}
RUN jira completion bash > /etc/bash_completion.d/jira

COPY --from=omc-builder       /${OUTPUT_DIR}/omc       ${BIN_DIR}
RUN ocm completion bash > /etc/bash_completion.d/omc

COPY --from=oc-nodepp-builder /${OUTPUT_DIR}/oc-nodepp ${BIN_DIR}
RUN oc-nodepp --help

# Install utils
COPY utils/bin /root/.local/bin

RUN omg completion bash > /etc/bash_completion.d/omg

# ssm plugin installed above; verify it
RUN /usr/local/aws-cli/aws ssm help

RUN git clone --depth 1 https://github.com/junegunn/fzf.git /root/.fzf \
    && /root/.fzf/install --all

# Setup bashrc.d directory
# Files with a ".bashrc" extension are sourced on login
COPY utils/bashrc.d /root/.bashrc.d
RUN printf 'if [ -d ${HOME}/.bashrc.d ] ; then\n  for file in ~/.bashrc.d/*.bashrc ; do\n    source ${file}\n  done\nfi\n' >> /root/.bashrc \
    && printf "[ -f ~/.fzf.bash ] && source ~/.fzf.bash" >> /root/.bashrc \
    # don't run automatically run commands when pasting from clipboard with a newline
    && printf "set enable-bracketed-paste on\n" >> /root/.inputrc

# Cleanup Home Dir
RUN rm -rf /root/anaconda* /root/original-ks.cfg /root/buildinfo

WORKDIR /root

ENV IO_OPENSHIFT_MANAGED_NAME="ocm-container"
ENV IO_OPENSHIFT_MANAGED_COMPONENT="full"
