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
      findutils \
      git \
      golang \
      google-cloud-cli \
      jq \
      make \
      nodejs \
      nodejs-nodemon \
      npm \
      openssl \
      procps-ng \
      python3 \
      python3-pip \
      rsync \
      sshuttle \
      tar \
      vim-enhanced \
      wget \
      && microdnf clean all \
      && rm -rf /var/cache/yum

RUN git clone --depth 1 https://github.com/junegunn/fzf.git /root/.fzf \
      && /root/.fzf/install --all

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

FROM builder as oc-builder
# Add `oc` to interact with openshift clusters (similar to kubectl)
# Replace version with a version number to pin a specific version (eg: "4.7.8")
ARG OC_VERSION="stable"
ENV OC_URL="https://mirror.openshift.com/pub/openshift-v4/@@PLATFORM@@/clients/ocp/${OC_VERSION}"

# Install the latest OC Binary from the mirror
RUN mkdir /oc
WORKDIR /oc
# Download the checksum
RUN curl -sSLf $(platform_convert $OC_URL --arm64 --x86_64)/sha256sum.txt -o sha256sum.txt
# Download the binary tarball
RUN /bin/bash -c "curl -sSLf -O $(platform_convert $OC_URL --arm64 --x86_64)/$(awk '/openshift-client-linux-([[:digit:]]+.)([[:digit:]]+.)([[:digit:]]+).tar.gz/{print $2}' sha256sum.txt)"
# Check the tarball and checksum match
RUN bash -c 'sha256sum --check <(awk "/openshift-client-linux-([[:digit:]]+.)([[:digit:]]+.)([[:digit:]]+).tar.gz/{print}" sha256sum.txt)'
RUN tar --extract --gunzip --no-same-owner --directory /out oc --file *.tar.gz
RUN chmod -R +x /out


FROM builder as ocm-builder
# Add `ocm` utility for interacting with the ocm-api
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
ARG OCM_VERSION="tags/v0.1.66"
ENV OCM_URL_SLUG="openshift-online/ocm-cli"
ENV OCM_URL="https://api.github.com/repos/${OCM_URL_SLUG}/releases/${OCM_VERSION}"

# Install ocm
# ocm is not in a tarball
RUN mkdir /ocm
WORKDIR /ocm

## arm-native
# Download the checksum
RUN [[ $(platform_convert "@@PLATFORM@@" --x86_64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-arm64.sha256")) | .browser_download_url') -o sha256sum.txt"
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --x86_64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-arm64$")) | .browser_download_url')"
# Check the binary and checksum match
RUN [[ $(platform_convert "@@PLATFORM@@" --x86_64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c 'sha256sum --check <( grep linux-arm64$  sha256sum.txt )'

## x86-native
# Download the checksum
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-amd64.sha256")) | .browser_download_url') -o sha256sum.txt"
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-amd64$")) | .browser_download_url')"
# Check the binary and checksum match
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c 'sha256sum --check <( grep linux-amd64$  sha256sum.txt )'

## x86-or-arm
RUN cp ocm* /out/ocm
RUN chmod -R +x /out


FROM builder as omc-builder
# Add `omc` utility to inspect must-gathers easily with 'oc' like commands
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG OMC_VERSION="tags/v2.5.0"
ENV OMC_URL_SLUG="gmeghnag/omc"
ENV OMC_URL="https://api.github.com/repos/${OMC_URL_SLUG}/releases/${OMC_VERSION}"

# Install omc
RUN mkdir /omc
WORKDIR /omc
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("checksums.txt")) | .browser_download_url') -o md5sum.txt"

# Download the binary
## x86-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_x86_64.tar.gz")) | .browser_download_url')"
## arm-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OMC_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_arm64.tar.gz")) | .browser_download_url')"

# Check the binary and checksum match
RUN bash -c 'md5sum --check <( grep $(platform_convert "Linux_@@PLATFORM@@.tar.gz" --x86_64 --arm64)  md5sum.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /out omc --file *.tar.gz
RUN chmod -R +x /out


FROM builder as osdctl-builder
# Add `osdctl` utility for common OSD commands
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG OSDCTL_VERSION="tags/v0.15.0"
ENV OSDCTL_URL_SLUG="openshift/osdctl"
ENV OSDCTL_URL="https://api.github.com/repos/${OSDCTL_URL_SLUG}/releases/${OSDCTL_VERSION}"

# Install osdctl
RUN mkdir /osdctl
WORKDIR /osdctl
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${OSDCTL_URL} -o - | jq -r '.assets[] | select(.name|test("sha256sum.txt")) | .browser_download_url') -o sha256sum.txt"

# Download the binary tarball
## x86-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OSDCTL_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_x86_64")) | .browser_download_url') "
# arm-native
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OSDCTL_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_arm64")) | .browser_download_url') "

# Check the tarball and checksum match
RUN bash -c 'sha256sum --check <( grep $(platform_convert "Linux_@@PLATFORM@@" --x86_64 --arm64)  sha256sum.txt )'
RUN tar --extract --gunzip --no-same-owner --directory /out osdctl --file *.tar.gz
RUN chmod -R +x /out


FROM builder as rosa-builder
# Add `rosa` utility for interacting with rosa clusters
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.1.4")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG ROSA_VERSION="tags/v1.2.16"
ENV ROSA_URL_SLUG="openshift/rosa"
ENV ROSA_URL="https://api.github.com/repos/${ROSA_URL_SLUG}/releases/${ROSA_VERSION}"

# Install ROSA
## There is not arm build for ROSA CLI yet.
RUN mkdir /rosa
WORKDIR /rosa
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${ROSA_URL} -o - | jq -r '.assets[] | select(.name|test("rosa-linux-amd64.sha256")) | .browser_download_url') -o sha256sum.txt"
# Download the binary
# NOTE: ROSA does a different type of sha256 setup (one per file) so the "$" below is necessary to select the binary filename
# correctly, and does not use a tarball
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${ROSA_URL} -o - | jq -r '.assets[] | select(.name|test("rosa-linux-amd64$")) | .browser_download_url') "
# Check the binary and checksum match
RUN bash -c 'sha256sum --check <( grep rosa-linux-amd64  sha256sum.txt )'
RUN mv rosa-linux-amd64 /out/rosa
RUN chmod -R +x /out

FROM builder as yq-builder
# Add `yq` utility for programatic yaml parsing
# the URL_SLUG is for checking the releasenotes when a version updates
ARG YQ_VERSION="tags/v4.31.1"
ENV YQ_URL_SLUG="mikefarah/yq"
ENV YQ_URL="https://api.github.com/repos/${YQ_URL_SLUG}/releases/${YQ_VERSION}"

# Install yq
RUN mkdir /yq
WORKDIR /yq
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("checksums$")) | .browser_download_url') -o checksums"
ENV LD_LIBRARY_PATH=/usr/local/lib

## amd64
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "curl -sSLf -o yq $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("linux_amd64$")) | .browser_download_url')"
# Check the binary and checksum match
# This is terrible, but not sure how to do this better.
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && exit 0 || /bin/bash -c "echo yq  $(awk '/yq_linux_amd64[[:space:]]/ {print $19}' checksums) | rhash -c -"

## arm64
# Download the binary
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || /bin/bash -c "curl -sSLf -o yq $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("linux_arm64$")) | .browser_download_url')"
# Check the binary and checksum match
# This is terrible, but not sure how to do this better.
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "arm64" ]] && exit 0 || echo "yq  $(awk '/yq_linux_arm64[[:space:]]/ {print $19}' checksums)" | rhash -c -

RUN cp yq /out/yq
RUN chmod -R +x /out


FROM builder as aws-builder
# Replace AWS client zipfile with specific file to pin to a specific version
# (eg: "awscli-exe-linux-x86_64-2.7.11.zip")
ARG AWSCLI_VERSION="awscli-exe-linux-@@PLATFORM@@.zip"
ENV AWSCLI_URL="https://awscli.amazonaws.com/${AWSCLI_VERSION}"
ENV AWSSIG_URL="https://awscli.amazonaws.com/${AWSCLI_VERSION}.sig"

# Install aws-cli
RUN mkdir -p /aws/bin
WORKDIR /aws
# Install the AWS CLI team GPG public key
COPY utils/dockerfile_assets/aws-cli.gpg ./
RUN gpg --import aws-cli.gpg
# Download the awscli GPG signature file
RUN curl -sSLf $(platform_convert $AWSSIG_URL --x86_64 --aarch64) -o awscliv2.zip.sig
# Download the awscli zip file
RUN curl -sSLf $(platform_convert $AWSCLI_URL --x86_64 --aarch64) -o awscliv2.zip
# Verify the awscli zip file
RUN gpg --verify awscliv2.zip.sig awscliv2.zip
# Extract the awscli zip
RUN unzip awscliv2.zip
# Install the libs to the usual location, so the simlinks will be right
# The final image build will copy them later
# Install the bins to the /aws/bin dir so the final image build copy is easier
RUN ./aws/install -b /aws/bin
RUN chmod -R +x /aws/bin


FROM builder as jira-builder
# Add `jira` utility for working with OHSS tickets
# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
# the URL_SLUG is for checking the releasenotes when a version updates
ARG JIRA_VERSION="tags/v1.3.0"
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


###########################
## Build the final image ##
###########################
# This is based on the first image build, with the yum packages installed
FROM dnf-install

# Copy previously acquired binaries into the $PATH
ENV BIN_DIR="/usr/local/bin"
COPY --from=aws-builder /aws/bin/ ${BIN_DIR}
COPY --from=aws-builder /usr/local/aws-cli /usr/local/aws-cli
COPY --from=jira-builder /out/jira ${BIN_DIR}
COPY --from=oc-builder /out/oc ${BIN_DIR}
COPY --from=ocm-builder /out/ocm ${BIN_DIR}
COPY --from=omc-builder /out/omc ${BIN_DIR}
COPY --from=osdctl-builder /out/osdctl ${BIN_DIR}
COPY --from=rosa-builder /out/rosa ${BIN_DIR}
COPY --from=yq-builder /out/yq ${BIN_DIR}

# Validate
RUN aws --version
RUN aws_completer bash > /etc/bash_completion.d/aws-cli
RUN jira completion bash > /etc/bash_completion.d/jira
RUN oc completion bash > /etc/bash_completion.d/oc
RUN ocm completion > /etc/bash_completion.d/ocm
RUN osdctl completion bash --skip-version-check > /etc/bash_completion.d/osdctl
RUN yq --version

# rosa is only available for amd64 platforms so ignore it
RUN [[ $(platform_convert "@@PLATFORM@@" --amd64 --arm64) != "amd64" ]] && rm ${BIN_DIR}/rosa || rosa completion bash > /etc/bash_completion.d/rosa

# Setup utils in $PATH
ENV PATH "$PATH:/root/.local/bin"

# Install utils
COPY utils/bin /root/.local/bin

# Install o-must-gather
# Replace "" with "=={tag}" to pin to a specific version (eg: "==1.2.6")
ARG O_MUST_GATHER_VERSION=""
RUN pip3 install --no-cache-dir o-must-gather${O_MUST_GATHER_VERSION}

# Setup pagerduty-cli
ARG PAGERDUTY_VERSION="0.1.11"
ENV HOME=/root
RUN npm install -g pagerduty-cli@${PAGERDUTY_VERSION}

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
