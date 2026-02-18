FROM debian:13 AS toolchain

ENV TZ=Asia/Seoul
ENV SHELL=/bin/bash

WORKDIR /var/tmp
COPY .devcontainer/scripts .
RUN ./install_base_packages.sh
RUN ./install_golang.sh --version 1.26.0
RUN ./install_delve_debugger.sh
RUN ./install_ai_assistants.sh

# 환경 설정
COPY .devcontainer/bashrc-settings /tmp/bashrc-settings
RUN cat /tmp/bashrc-settings > /root/.bashrc && \
    rm /tmp/bashrc-settings && \
    echo "set encoding=utf-8" > /root/.vimrc && \
    echo "set mouse=" >> /root/.vimrc && \
    localedef -f UTF-8 -i ko_KR ko_KR.UTF-8 && \
    localedef -f UTF-8 -i en_US en_US.UTF-8

WORKDIR /workspace