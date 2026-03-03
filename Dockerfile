# Toolchain image to build Go.
FROM debian:13 AS toolchain

SHELL ["/bin/bash", "-lc"]

ENV TZ=Asia/Seoul
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
# WORKSPACE_ROOT is the path of the container workspace. The default value should
# be specified for manual Docker build process, but it will be overridden by the 
# devcontainer configuration.
ARG WORKSPACE_ROOT=/workspace
ENV WORKSPACE_ROOT=${WORKSPACE_ROOT}

WORKDIR /var/tmp/scripts
COPY .devcontainer/scripts/install_base_packages.sh .
COPY .devcontainer/scripts/install_golang.sh .
COPY .devcontainer/scripts/install_golang_extra_packages.sh .
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

# Builder image to pass a compiled binary to the final runtime image.
FROM toolchain AS builder

WORKDIR ${WORKSPACE_ROOT}
COPY . .
RUN make build \
    && mkdir -p /app \
    && mv bear-go /app

# Devcontainer image for Go development. This image may be used in both local 
# development and GitHub Actions, so additional packages for local development 
# are installed conditionally.
FROM toolchain AS dev

# GITHUB_ACTIONS is set to true when the image is built in GitHub Actions, and 
# it is used to determine whether to install additional packages for local 
# development or not.
ARG GITHUB_ACTIONS
ENV GITHUB_ACTIONS=${GITHUB_ACTIONS}

# Install additional packages for local development if not running in GitHub Actions.
WORKDIR /var/tmp/scripts
COPY .devcontainer/scripts/install_ai_assistants.sh .
COPY .devcontainer/scripts/install_watchexec.sh .
RUN if [ "${GITHUB_ACTIONS}" != "true" ]; then \
        chmod +x ./*.sh; \
        ./install_ai_assistants.sh; \
        ./install_watchexec.sh; \
    fi

WORKDIR ${WORKSPACE_ROOT}

# Final runtime image to deploy the compiled binary.
FROM dhi.io/static:20250419-glibc-debian13 AS runtime

COPY --from=builder /app/bear-go /app/bear-go
ENTRYPOINT ["/app/bear-go"]