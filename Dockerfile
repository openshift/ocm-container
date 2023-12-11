### Pre-install yum stuff
ARG BASE_IMAGE=registry.access.redhat.com/ubi9/ubi-minimal:9.1.0
FROM ${BASE_IMAGE} as base-update

RUN microdnf --assumeyes --nodocs update \
      && microdnf clean all \
      && rm -rf /var/cache/yum

FROM base-update as dnf-install

# OCM backplane console port to map
ENV OCM_BACKPLANE_CONSOLE_PORT 9999

# Adds Platform Conversion Tool for arm64/x86_64 compatibility
COPY utils/dockerfile_assets/platforms.sh /usr/local/bin/platform_convert

# Add google repo
COPY utils/dockerfile_assets/google-cloud-sdk.repo /etc/yum.repos.d/
RUN platform_convert -i /etc/yum.repos.d/google-cloud-sdk.repo --x86_64 --aarch64

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
      google-cloud-cli \
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
      sshuttle \
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

### Download the binaries
# Anything in this image must be COPY'd into the final image, below
FROM ${BASE_IMAGE} as builder

# Adds Platform Conversion Tool for arm64/x86_64 compatibility
# need to add this a second time to add it to the builder image
COPY utils/dockerfile_assets/platforms.sh /usr/local/bin/platform_convert

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

# Directory for the extracted binaries, etc
RUN mkdir -p /out

##############################
## Individual Binary Builds ##
##############################

FROM builder as omc-builder
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
RUN tar --extract --gunzip --no-same-owner --directory /out omc --file *.tar.gz
RUN chmod -R +x /out

FROM builder as jira-builder
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
RUN tar --extract --gunzip --no-same-owner --directory /out --strip-components=2 */bin/jira --file *.tar.gz
RUN chmod -R +x /out

FROM builder as k9s-builder
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
RUN tar --extract --gunzip --no-same-owner --directory /out k9s --file *.tar.gz
RUN chmod +x /out/k9s

FROM builder as oc-nodepp-builder
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
RUN tar --extract --gunzip --no-same-owner --directory /out oc-nodepp --file *.tar.gz
RUN chmod +x /out/oc-nodepp

FROM builder as backplane-tools-builder
# Install via backplane-tools
ARG BACKPLANE_TOOLS_VERSION="tags/v0.4.0"
ENV BACKPLANE_TOOLS_URL_SLUG="openshift/backplane-tools"
ENV BACKPLANE_TOOLS_URL="https://api.github.com/repos/${BACKPLANE_TOOLS_URL_SLUG}/releases/${BACKPLANE_TOOLS_VERSION}"
RUN mkdir /backplane-tools
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
ENV PATH "$PATH:/root/.local/bin/backplane/latest"
RUN /usr/local/bin/backplane-tools install all

# Copy symlink sources from ./local/bin to /out
RUN cp -Hv  /root/.local/bin/backplane/latest/* /out
RUN chmod +x /out/*
# copy aws cli assets
RUN cp -r /root/.local/bin/backplane/aws/*/aws-cli/dist /out/aws_dist

# Copy hypershift binary
FROM quay.io/acm-d/rhtap-hypershift-operator as hypershift
RUN mkdir -p /out
RUN cp /usr/bin/hypershift /out/hypershift
RUN chmod -R +x /out

###########################
## Build the final image ##
###########################
# This is based on the first image build, with the yum packages installed
FROM dnf-install
ENV BIN_DIR="/usr/local/bin"

# Copy previously acquired binaries into the $PATH
WORKDIR /
COPY --from=jira-builder /out/jira ${BIN_DIR}
COPY --from=omc-builder /out/omc ${BIN_DIR}
COPY --from=k9s-builder /out/k9s ${BIN_DIR}
COPY --from=oc-nodepp-builder /out/oc-nodepp ${BIN_DIR}
COPY --from=backplane-tools-builder /out/oc ${BIN_DIR}
COPY --from=backplane-tools-builder /out/ocm ${BIN_DIR}
COPY --from=backplane-tools-builder /out/ocm-backplane ${BIN_DIR}
COPY --from=backplane-tools-builder /out/osdctl ${BIN_DIR}
COPY --from=backplane-tools-builder /out/rosa ${BIN_DIR}
COPY --from=backplane-tools-builder /out/yq ${BIN_DIR}
COPY --from=backplane-tools-builder /out/aws_dist /usr/local/aws-cli
COPY --from=hypershift /out/hypershift ${BIN_DIR}

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
RUN ocm backplane completion bash > /etc/bash_completion.d/ocm-backplane
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && echo "removing non-arm64 hypershift binary" && rm ${BIN_DIR}/hypershift || hypershift --version
RUN rosa completion bash > /etc/bash_completion.d/rosa

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

# Set an exposable port for the cluster console proxy
# Can be used with `-o "-P"` to map 9999 inside the container to a random port at runtime
EXPOSE $OCM_BACKPLANE_CONSOLE_PORT

WORKDIR /root
ENTRYPOINT ["/bin/bash"]
