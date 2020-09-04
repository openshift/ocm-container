#FROM ocm-container
FROM fedora:latest

ARG osv4client=openshift-client-linux-4.3.12.tar.gz
ENV osv4client=$osv4client

ADD ./container-setup /container-setup

WORKDIR /container-setup

RUN ./install.sh I-am-in-container

