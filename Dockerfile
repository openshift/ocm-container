### Pre-install yum stuff
FROM fedora:latest as dnf-install

# Install packages
# These packages will end up in the final image
# Installed here to save build time
RUN yum --assumeyes install \
    bash-completion \
    findutils \
    fzf \
    jq \
    npm \
    make \
    openssl \
    procps-ng \
    python-pip \
    rsync \
    sshuttle \
    the_silver_searcher \
    tmux \
    vim-enhanced \
    wget \
    && yum clean all;


### Download the binaries
# Anything in this image must be COPY'd into the final image, below
FROM fedora:latest as builder

# jq is a pre-req for making parsing of download urls easier
RUN dnf install -y jq rhash unzip

# Replace version with a version number to pin a specific version (eg: "4.7.8")
ARG OC_VERSION="stable"
ENV OC_URL="https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/${OC_VERSION}"

# Replace version with a version number to pin a specific version (eg: "4.7.8")
ARG ROSA_VERSION="latest"
ENV ROSA_URL="https://api.github.com/repos/openshift/rosa/releases/${ROSA_VERSION}"

# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
ARG OSDCTL_VERSION="latest"
ENV OSDCTL_URL="https://api.github.com/repos/openshift/osdctl/releases/${OSDCTL_VERSION}"

# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
ARG OCM_VERSION="latest"
ENV OCM_URL="https://api.github.com/repos/openshift-online/ocm-cli/releases/${OCM_VERSION}"

# Replace "/latest" with "/tags/{tag}" to pin to a specific version (eg: "/tags/v0.4.0")
ARG VELERO_VERSION="latest"
ENV VELERO_URL="https://api.github.com/repos/vmware-tanzu/velero/releases/${VELERO_VERSION}"

# Replace AWS client zipfile with specific file to pin to a specific version
# (eg: "awscli-exe-linux-x86_64-2.0.30.zip")
ARG AWSCLI_VERSION="awscli-exe-linux-x86_64.zip"
ENV AWSCLI_URL="https://awscli.amazonaws.com/${AWSCLI_VERSION}"
ENV AWSSIG_URL="https://awscli.amazonaws.com/${AWSCLI_VERSION}.sig"

# Add `yq` utility for programatic yaml parsing
ARG YQ_VERSION="latest"
ENV YQ_URL="https://api.github.com/repos/mikefarah/yq/releases/${YQ_VERSION}"

# Directory for the extracted binaries, etc
RUN mkdir -p /out

# Install the latest OC Binary from the mirror
RUN mkdir /oc
WORKDIR /oc
# Download the checksum
RUN curl -sSLf ${OC_URL}/sha256sum.txt -o sha256sum.txt
# Download the binary tarball
RUN /bin/bash -c "curl -sSLf -O ${OC_URL}/$(awk -v asset="openshift-client-linux" '$0~asset {print $2}' sha256sum.txt)"
# Check the tarball and checksum match
RUN sha256sum --check --ignore-missing sha256sum.txt
RUN tar --extract --gunzip --no-same-owner --directory /out oc --file *.tar.gz

# Install ROSA
RUN mkdir /rosa
WORKDIR /rosa
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${ROSA_URL} -o - | jq -r '.assets[] | select(.name|test("rosa-linux-amd64.sha256")) | .browser_download_url') -o sha256sum.txt"
# Download the binary
# NOTE: ROSA does a different type of sha256 setup (one per file) so the "$" below is necessary to select the binary filename
# correctly, and does not use a tarball
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${ROSA_URL} -o - | jq -r '.assets[] | select(.name|test("rosa-linux-amd64$")) | .browser_download_url') "
# Check the binary and checksum match
RUN sha256sum --check --ignore-missing sha256sum.txt
RUN mv rosa-linux-amd64 /out/rosa

# Install osdctl
RUN mkdir /osdctl
WORKDIR /osdctl
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${OSDCTL_URL} -o - | jq -r '.assets[] | select(.name|test("sha256sum.txt")) | .browser_download_url') -o sha256sum.txt"
# Download the binary tarball
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OSDCTL_URL} -o - | jq -r '.assets[] | select(.name|test("Linux_x86_64")) | .browser_download_url') "
# Check the tarball and checksum match
RUN sha256sum --check --ignore-missing sha256sum.txt
RUN tar --extract --gunzip --no-same-owner --directory /out osdctl --file *.tar.gz

# Install ocm
# ocm is not in a tarball
RUN mkdir /ocm
WORKDIR /ocm
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-amd64.sha256")) | .browser_download_url') -o sha256sum.txt"
# Download the binary
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${OCM_URL} -o - | jq -r '.assets[] | select(.name|test("linux-amd64$")) | .browser_download_url')"
# Check the binary and checksum match
RUN sha256sum --check --ignore-missing sha256sum.txt
RUN cp ocm* /out/ocm

# Install velero
RUN mkdir /velero
WORKDIR /velero
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${VELERO_URL} -o - | jq -r '.assets[] | select(.name|test("CHECKSUM")) | .browser_download_url') -o sha256sum.txt"
# Download the binary tarball
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${VELERO_URL} -o - | jq -r '.assets[] | select(.name|test("linux-amd64")) | .browser_download_url') "
# Check the tarball and checksum match
RUN sha256sum --check --ignore-missing sha256sum.txt
RUN tar --extract --gunzip --no-same-owner --directory /out --wildcards --no-wildcards-match-slash --no-anchored --strip-components=1 *velero --file *.tar.gz

# Install yq
RUN mkdir /yq
WORKDIR /yq
# Download the checksum
RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("checksums$")) | .browser_download_url') -o checksums"
# RUN /bin/bash -c "curl -sSLf $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("checksums_hashes_order")) | .browser_download_url') -o checksums_hashes_order"
# Download the binary
RUN /bin/bash -c "curl -sSLf -O $(curl -sSLf ${YQ_URL} -o - | jq -r '.assets[] | select(.name|test("linux_amd64$")) | .browser_download_url') "
# Check the binary and checksum match
# This is terrible, but not sure how to do this better.
RUN rhash -c checksums | grep Success:1
RUN cp yq_linux_amd64 /out/yq

# Install aws-cli
RUN mkdir -p /aws/bin
WORKDIR /aws
# Install the AWS CLI team GPG public key
COPY aws-cli.gpg ./
RUN gpg --import aws-cli.gpg
# Download the awscli GPG signature file
RUN curl -sSLf $AWSSIG_URL -o awscliv2.zip.sig
# Download the awscli zip file
RUN curl -sSLf $AWSCLI_URL -o awscliv2.zip
# Verify the awscli zip file
RUN gpg --verify awscliv2.zip.sig awscliv2.zip
# Extract the awscli zip
RUN unzip awscliv2.zip
# Install the libs to the usual location, so the simlinks will be right
# The final image build will copy them later
# Install the bins to the /aws/bin dir so the final image build copy is easier
RUN ./aws/install -b /aws/bin

# Make binaries executable
RUN chmod -R +x /out

### Build the final image
# This is based on the first image build, with the yum packages installed
FROM dnf-install

# Copy previously acquired binaries into the $PATH
ENV BIN_DIR="/usr/local/bin"
COPY --from=builder /out/oc ${BIN_DIR}
COPY --from=builder /out/rosa ${BIN_DIR}
COPY --from=builder /out/osdctl ${BIN_DIR}
COPY --from=builder /out/ocm ${BIN_DIR}
COPY --from=builder /out/velero ${BIN_DIR}
COPY --from=builder /aws/bin/ ${BIN_DIR}
COPY --from=builder /usr/local/aws-cli /usr/local/aws-cli
COPY --from=builder /out/yq ${BIN_DIR}

# Validate
RUN oc completion bash > /etc/bash_completion.d/oc
RUN rosa completion bash > /etc/bash_completion.d/rosa
RUN osdctl completion bash > /etc/bash_completion.d/osdctl
RUN ocm completion > /etc/bash_completion.d/ocm
RUN velero completion bash > /etc/bash_completion.d/velero
RUN aws_completer bash > /etc/bash_completion.d/aws-cli
RUN aws --version
RUN yq --version

# Setup utils in $PATH
ENV PATH "$PATH:/root/.local/bin"

# Install utils
COPY utils/bin /root/.local/bin

# Install o-must-gather
RUN pip3 install o-must-gather

# Setup pagerduty-cli
RUN npm install -g pagerduty-cli@0.0.70

# Setup bashrc.d directory
# Files with a ".bashrc" extension are sourced on login
COPY utils/bashrc.d /root/.bashrc.d
RUN printf 'if [ -d ${HOME}/.bashrc.d ] ; then\n  for file in ~/.bashrc.d/*.bashrc ; do\n    source ${file}\n  done\nfi\n' >> /root/.bashrc

# Cleanup Home Dir
RUN rm /root/anaconda* /root/original-ks.cfg

WORKDIR /root
ENTRYPOINT ["/bin/bash"]
