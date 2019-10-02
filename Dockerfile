#FROM ocm-container
FROM fedora:latest

ADD ./container-setup /container-setup

RUN /container-setup/install.sh I-am-in-container
