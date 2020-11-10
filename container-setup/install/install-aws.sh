#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

#export awsclient=awscli-exe-linux-x86_64.zip

mkdir /usr/local/aws;
pushd /usr/local/aws;
wget -q https://awscli.amazonaws.com/${awsclient}
unzip ${awsclient}
rm ${awsclient}
./aws/install;
popd;
