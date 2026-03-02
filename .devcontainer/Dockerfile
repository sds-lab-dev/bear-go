# Toolchain image to build Go.
FROM debian:13

ENV TZ=Asia/Seoul
ENV SHELL=/bin/bash
ENV IS_SANDBOX=1
ENV ENABLE_LSP_TOOL=1
ENV CLAUDE_CONFIG_DIR=/root/.persist/claude
ENV CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
ENV CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY=1
ENV CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1
ENV XDG_CONFIG_HOME=/root/.persist/xdg/config
ENV XDG_CACHE_HOME=/root/.persist/xdg/cache
ENV XDG_DATA_HOME=/root/.persist/xdg/data
ENV GOPATH=/root/.persist/go
ENV GOMODCACHE=/root/.persist/go/pkg/mod
ENV GOCACHE=/root/.persist/go/build-cache
ENV GOENV=/root/.persist/go/env
# GITHUB_ACTIONS is set to true when the image is built in GitHub Actions, and 
# it is used to determine whether to install additional packages for local 
# development or not.
ARG GITHUB_ACTIONS
ENV GITHUB_ACTIONS=${GITHUB_ACTIONS}
# WORKSPACE_ROOT is the path of the container workspace.
ARG WORKSPACE_ROOT
ENV WORKSPACE_ROOT=${WORKSPACE_ROOT}

WORKDIR /var/tmp/scripts
COPY .devcontainer/scripts .
RUN chmod +x ./*.sh \
    && ./install_base_packages.sh \
    && ./install_golang.sh --version 1.26.0 \
    && ./install_golang_extra_packages.sh
# Install additional packages for local development if not running in GitHub Actions.
RUN if [ "${GITHUB_ACTIONS}" != "true" ]; then \
        ./install_ai_assistants.sh; \
        ./install_watchexec.sh; \
    fi

WORKDIR /root
COPY .devcontainer/bashrc-settings /tmp/bashrc-settings
RUN cat /tmp/bashrc-settings > .bashrc \
    && rm /tmp/bashrc-settings \
    && echo "set encoding=utf-8" > .vimrc \
    && echo "set mouse=" >> .vimrc \
    && localedef -f UTF-8 -i ko_KR ko_KR.UTF-8 \
    && localedef -f UTF-8 -i en_US en_US.UTF-8

WORKDIR ${WORKSPACE_ROOT}