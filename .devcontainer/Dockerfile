# Toolchain image to build Go.
FROM debian:13 AS toolchain

ENV TZ=Asia/Seoul
ENV SHELL=/bin/bash

WORKDIR /var/tmp/scripts
COPY .devcontainer/scripts .
RUN chmod +x ./*.sh \
    && ./install_base_packages.sh \
    && ./install_golang.sh --version 1.26.0 \
    && ./install_golang_extra_packages.sh

WORKDIR /root
COPY .devcontainer/bashrc-settings /tmp/bashrc-settings
RUN cat /tmp/bashrc-settings > .bashrc \
    && rm /tmp/bashrc-settings \
    && echo "set encoding=utf-8" > .vimrc \
    && echo "set mouse=" >> .vimrc \
    && localedef -f UTF-8 -i ko_KR ko_KR.UTF-8 \
    && localedef -f UTF-8 -i en_US en_US.UTF-8

# Devcontainer image to use for local development, which includes the toolchain 
# and other necessary tools.
FROM toolchain AS dev

ARG WORKSPACE_ROOT
ENV WORKSPACE_ROOT=${WORKSPACE_ROOT}

WORKDIR /var/tmp/scripts
COPY .devcontainer/scripts .
RUN chmod +x ./*.sh \
    && ./install_ai_assistants.sh \
    && ./install_watchexec.sh

WORKDIR ${WORKSPACE_ROOT}