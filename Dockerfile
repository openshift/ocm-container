#FROM ocm-container
FROM fedora:latest

ARG osv4client=openshift-client-linux-4.3.5.tar.gz
ENV osv4client=$osv4client

ADD ./container-setup /container-setup

RUN /container-setup/install.sh I-am-in-container
