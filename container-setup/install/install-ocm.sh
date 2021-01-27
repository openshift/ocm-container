#!/bin/bash -e

set -o errexit

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";
source /container-setup/install/helpers.sh

LATEST_RELEASE="https://github.com/openshift-online/ocm-cli/releases/latest/download/ocm-linux-amd64"

mkdir /usr/local/ocm \
  && pushd /usr/local/ocm

echo "Downloading the latest release"
curl -sSL ${LATEST_RELEASE} -o ocm-linux-amd64
echo "Validating binary against SHA256 sum"
curl -sSL ${LATEST_RELEASE}.sha256 | sha256sum --check --status
echo "Making binary executable"
chmod +x ocm-linux-amd64
echo "Symlinking binary to \"ocm\""
ln -s /usr/local/ocm/ocm-linux-amd64 /usr/local/bin/ocm
echo "Checking binary execution"
ocm version

popd

ocm completion > /etc/bash_completion.d/ocm
