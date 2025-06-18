ARG BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.6-1749489516
ARG HYPERSHIFT_BASE_IMAGE=quay.io/hypershift/hypershift-operator
FROM ${BASE_IMAGE} as tools-base
ARG OUTPUT_DIR="/opt"

RUN microdnf --assumeyes install gzip jq tar

# Adds Platform Conversion Tool for arm64/x86_64 compatibility
# need to add this a second time to add it to the builder image
COPY utils/dockerfile_assets/platforms.sh /usr/local/bin/platform_convert

### BACKPLANE TOOLS - download SRE standad binaries to a temporary container
FROM tools-base as backplane-tools
ARG OUTPUT_DIR="/opt"

# Set GH_TOKEN to use authenticated GH requests
ARG GITHUB_TOKEN

ARG BACKPLANE_TOOLS_VERSION="tags/v1.2.0"
ENV BACKPLANE_TOOLS_URL_SLUG="openshift/backplane-tools"
ENV BACKPLANE_TOOLS_URL="https://api.github.com/repos/${BACKPLANE_TOOLS_URL_SLUG}/releases/${BACKPLANE_TOOLS_VERSION}"
ENV BACKPLANE_BIN_DIR="/root/.local/bin/backplane"

RUN mkdir -p /backplane-tools
WORKDIR /backplane-tools

# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${BACKPLANE_TOOLS_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.txt")) | .browser_download_url') -o checksums.txt"

# Download amd64 binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${BACKPLANE_TOOLS_URL} -o - | jq -r '.assets[] | select(.name|test("linux_amd64")) | .browser_download_url') "
# Download arm64 binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${BACKPLANE_TOOLS_URL} -o - | jq -r '.assets[] | select(.name|test("linux_arm64")) | .browser_download_url') "

# Extract
RUN tar --extract --gunzip --no-same-owner --directory "/usr/local/bin"  --file *.tar.gz

# Install all using backplane-tools
RUN /bin/bash -c "PATH=${PATH}:${BACKPLANE_BIN_DIR}/latest /usr/local/bin/backplane-tools install all"

# Copy symlink sources from ./local/bin to $OUTPUT_DIR
RUN cp -Hv  ${BACKPLANE_BIN_DIR}/latest/* ${OUTPUT_DIR}

# copy aws cli assets
RUN cp -r ${BACKPLANE_BIN_DIR}/aws/*/aws-cli/dist /${OUTPUT_DIR}/aws_dist

# Copy hypershift binary
FROM ${HYPERSHIFT_BASE_IMAGE} AS hypershift
ARG OUTPUT_DIR="/opt"
RUN cp /usr/bin/hypershift /${OUTPUT_DIR}/hypershift

### Builder - Get or Build Individual Binaries
FROM tools-base as builder
ARG OUTPUT_DIR="/opt"

# jq is a pre-req for making parsing of download urls easier
RUN microdnf --assumeyes --nodocs install \
      gcc \
      git \
      jq \
      make \
      tar \
      unzip

# install from epel
RUN curl -sSlo epel-gpg https://dl.fedoraproject.org/pub/epel/RPM-GPG-KEY-EPEL-9 \
      && rpm --import epel-gpg \
      && rpm -ivh https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm \
      && microdnf --assumeyes --nodocs install rhash

# Directory for the extracted binaries, etc; used in child images
RUN mkdir -p /${OUTPUT_DIR}

FROM builder as omc-builder
ARG OUTPUT_DIR="/opt"
# Add `omc` utility to inspect must-gathers easily with 'oc' like commands
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG OMC_VERSION="tags/v3.3.2"
ENV OMC_URL_SLUG="gmeghnag/omc"
ENV OMC_URL="https://api.github.com/repos/${OMC_URL_SLUG}/releases/${OMC_VERSION}"

# Install omc
RUN mkdir /omc
WORKDIR /omc
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.txt")) | .browser_download_url') -o md5sum.txt"

# Download the binary
# x86-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_x86_64.tar.gz")) | .browser_download_url')"
# arm-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_arm64.tar.gz")) | .browser_download_url')"

# Check the binary and checksum match
RUN bash -c 'md5sum --check <( grep $(platform_convert "Linux_@@PLATFORM@@.tar.gz" --x86_64 --arm64)  md5sum.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} omc --file *.tar.gz
RUN chmod -R +x /${OUTPUT_DIR}

FROM builder as jira-builder
ARG OUTPUT_DIR="/opt"
# Add `jira` utility for working with OHSS tickets
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG JIRA_VERSION="tags/v1.4.0"
ENV JIRA_URL_SLUG="ankitpokhrel/jira-cli"
ENV JIRA_URL="https://api.github.com/repos/${JIRA_URL_SLUG}/releases/${JIRA_VERSION}"
WORKDIR /jira
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${JIRA_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.txt")) | .browser_download_url') -o checksums.txt"

## amd64
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${JIRA_URL} -o - | jq -r '.assets[] | select(.name|test("linux_x86_64")) | .browser_download_url') "
## arm64
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${JIRA_URL} -o - | jq -r '.assets[] | select(.name|test("linux_arm64")) | .browser_download_url') "

# Check the tarball and checksum match
RUN bash -c 'sha256sum --check <( grep $(platform_convert "linux_@@PLATFORM@@" --x86_64 --arm64)  checksums.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} --strip-components=2 */bin/jira --file *.tar.gz
RUN chmod -R +x /${OUTPUT_DIR}

FROM builder as k9s-builder
ARG OUTPUT_DIR="/opt"
# Add `k9s` utility
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG K9S_VERSION="latest"
ENV K9S_URL_SLUG="derailed/k9s"
ENV K9S_URL="https://api.github.com/repos/${K9S_URL_SLUG}/releases/${K9S_VERSION}"

# Install k9s
RUN mkdir /k9s
WORKDIR /k9s

# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${K9S_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.sha256")) | .browser_download_url') -o sha256sum.txt"

# Download the binary tarball
# x86-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${K9S_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_amd64.tar.gz$")) | .browser_download_url')"
# arm-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${K9S_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_arm64.tar.gz$")) | .browser_download_url')"

# Check the tarball and checksum match
RUN bash -c 'sha256sum --check <( grep $(platform_convert "Linux_@@PLATFORM@@.tar.gz$" --amd64 --arm64)  sha256sum.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} k9s --file *.tar.gz
RUN chmod +x /${OUTPUT_DIR}/k9s

FROM builder as oc-nodepp-builder
ARG OUTPUT_DIR="/opt"
# Add `oc-nodepp` utility
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG NODEPP_VERSION="tags/v0.1.2"
ENV NODEPP_URL_SLUG="mrbarge/oc-nodepp"
ENV NODEPP_URL="https://api.github.com/repos/${NODEPP_URL_SLUG}/releases/${NODEPP_VERSION}"
# Install oc-nodepp
RUN mkdir /nodepp
WORKDIR /nodepp

# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${NODEPP_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.txt")) | .browser_download_url') -o sha256sum.txt"

# Download the binary tarball
# x86-native
RUN [[ $(platform_convert "@@PLATFORM@@" --x86_64 --arm64) != "x86_64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${NODEPP_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_x86_64")) | .browser_download_url') "
# arm-native
RUN [[ $(platform_convert "@@PLATFORM@@" --x86_64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${NODEPP_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_arm64")) | .browser_download_url') "

# Check the tarball and checksum match
RUN bash -c 'sha256sum --check <( grep $(platform_convert "Linux_@@PLATFORM@@.tar.gz" --x86_64 --arm64)  sha256sum.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /${OUTPUT_DIR} oc-nodepp --file *.tar.gz
RUN chmod +x /${OUTPUT_DIR}/oc-nodepp

### Pre-install yum stuff for final images
FROM ${BASE_IMAGE} as base-update
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"

RUN microdnf --assumeyes install yum-utils \
      && microdnf --assumeyes --nodocs update \
      && microdnf clean all \
      && rm -rf /var/cache/yum

ENV IO_OPENSHIFT_MANAGED_NAME="ocm-container"
LABEL io.openshift.managed.name="ocm-container"
LABEL io.openshift.managed.description="Containerized environment for accessing OpenShift v4 clusters, packing necessary tools/scripts"

# Set an exposable port for the cluster console proxy
# Can be used with `-o "-P"` to map 9999 inside the container to a random port at runtime
ENV OCM_BACKPLANE_CONSOLE_PORT 9999
EXPOSE $OCM_BACKPLANE_CONSOLE_PORT
ENTRYPOINT ["/bin/bash"]

# Create a directory for the ocm config file
RUN mkdir -p /root/.config/ocm

### Final Minimal Image
FROM base-update as ocm-container-minimal
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"
ARG BIN_DIR="/usr/local/bin"

COPY --from=backplane-tools /${OUTPUT_DIR}/aws_dist      /usr/local/aws-cli
COPY --from=backplane-tools /${OUTPUT_DIR}/oc            ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/ocm           ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/ocm-backplane ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/ocm-addons    ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/osdctl        ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/rosa          ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/servicelogger ${BIN_DIR}
COPY --from=backplane-tools /${OUTPUT_DIR}/yq            ${BIN_DIR}
COPY --from=hypershift      /${OUTPUT_DIR}/hypershift    ${BIN_DIR}

### DNF Install other tools on top of Minimal
FROM ocm-container-minimal as dnf-install

# Add Platform Conversion Tool for arm64/x86_64 compatibility
COPY utils/dockerfile_assets/platforms.sh /usr/local/bin/platform_convert

# Add epel repos
RUN rpm --import https://dl.fedoraproject.org/pub/epel/RPM-GPG-KEY-EPEL-9 \
      && rpm -ivh https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm

# Install packages
# These packages will end up in the final image
# Installed here to save build time
RUN microdnf --assumeyes --nodocs install \
      bash-completion \
      bind-utils \
      crun\
      findutils \
      fuse-overlayfs \
      git \
      golang \
      jq \
      make \
      nodejs \
      nodejs-nodemon \
      npm \
      openssl \
      podman \
      procps-ng \
      python3 \
      python3-pip \
      rsync \
      tar \
      vim-enhanced \
      wget \
      xz \
      && microdnf clean all \
      && rm -rf /var/cache/yum

RUN git clone --depth 1 https://github.com/junegunn/fzf.git /root/.fzf \
      && /root/.fzf/install --all

### podman container config
# Overlay over overlay is often denied by the kernel, so this creates non overlay volumes to be used within the container.
VOLUME /var/lib/containers

# copy storage.conf to enable fuse-overlayfs storage.
COPY utils/dockerfile_assets/storage.conf /etc/containers/storage.conf

# add containers.conf file to make sure containers run easier.
COPY utils/dockerfile_assets/containers.conf /etc/containers/containers.conf

###########################
## Build the final image ##
###########################
# This is based on the first image build, with the yum packages installed
FROM dnf-install as ocm-container
# ARG keeps the values from the final image
ARG OUTPUT_DIR="/opt"
ARG BIN_DIR="/usr/local/bin"

# Copy previously acquired binaries into the $PATH
WORKDIR /
COPY --from=jira-builder      /${OUTPUT_DIR}/jira      ${BIN_DIR}
COPY --from=omc-builder       /${OUTPUT_DIR}/omc       ${BIN_DIR}
COPY --from=k9s-builder       /${OUTPUT_DIR}/k9s       ${BIN_DIR}
COPY --from=oc-nodepp-builder /${OUTPUT_DIR}/oc-nodepp ${BIN_DIR}

# Validate
RUN /usr/local/aws-cli/aws --version
RUN /usr/local/aws-cli/aws_completer bash > /etc/bash_completion.d/aws-cli
RUN jira completion bash > /etc/bash_completion.d/jira
RUN oc completion bash > /etc/bash_completion.d/oc
RUN ocm completion > /etc/bash_completion.d/ocm
RUN osdctl completion bash --skip-version-check > /etc/bash_completion.d/osdctl
RUN yq --version
RUN k9s completion bash > /etc/bash_completion.d/k9s
RUN ocm backplane version
RUN ocm addons version
RUN ocm backplane completion bash > /etc/bash_completion.d/ocm-backplane
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && echo "removing non-arm64 hypershift binary" && rm ${BIN_DIR}/hypershift || hypershift --version 2> /dev/null
RUN rosa completion bash > /etc/bash_completion.d/rosa
RUN servicelogger version
RUN servicelogger completion bash > /etc/bash_completion.d/servicelogger

# Install utils
COPY utils/bin /root/.local/bin

# Install o-must-gather
# Replace "" with "=={tag}" to pin to a specific version (eg: "==1.2.6")
ARG O_MUST_GATHER_VERSION=""
RUN pip3 install --no-cache-dir o-must-gather${O_MUST_GATHER_VERSION}

# Setup pagerduty-cli
ARG PAGERDUTY_VERSION="0.1.18"
ENV HOME=/root
RUN npm install -g pagerduty-cli@${PAGERDUTY_VERSION}

# install ssm plugin
RUN rpm -i $(platform_convert https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_@@PLATFORM@@/session-manager-plugin.rpm --arm64 --custom-amd64 64bit)

# Setup bashrc.d directory
# Files with a ".bashrc" extension are sourced on login
COPY utils/bashrc.d /root/.bashrc.d
RUN printf 'if [ -d ${HOME}/.bashrc.d ] ; then\n  for file in ~/.bashrc.d/*.bashrc ; do\n    source ${file}\n  done\nfi\n' >> /root/.bashrc \
    && printf "[ -f ~/.fzf.bash ] && source ~/.fzf.bash" >> /root/.bashrc \
    # Setup pdcli autocomplete \
    &&  printf 'eval $(pd autocomplete:script bash)' >> /root/.bashrc.d/99-pdcli.bashrc \
    && bash -c "SHELL=/bin/bash pd autocomplete --refresh-cache" \
    # don't run automatically run commands when pasting from clipboard with a newline
    && printf "set enable-bracketed-paste on\n" >> /root/.inputrc

# Cleanup Home Dir
RUN rm -rf /root/anaconda* /root/original-ks.cfg /root/buildinfo

WORKDIR /root
