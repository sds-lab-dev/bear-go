# Toolchain image to build Go.
FROM debian:13 AS toolchain

SHELL ["/bin/bash", "-c"]

ENV TZ=Asia/Seoul

# WORKSPACE_ROOT is the path of the container workspace. The default value 
# should be specified for manual Docker build process, but it will be 
# overridden by the devcontainer configuration.
ARG WORKSPACE_ROOT=/workspace
ENV WORKSPACE_ROOT=${WORKSPACE_ROOT}

# GOPATH_VOLUME_DIR is the path of the Docker volume to persist Go compiled artifacts to speed up
# subsequent builds. If it is not set, use the default path inside the container, which does 
# not persist across restarts.
ARG GOPATH_VOLUME_DIR=/var/local/go
ENV GOPATH=${GOPATH_VOLUME_DIR}
ENV GOMODCACHE=${GOPATH_VOLUME_DIR}/pkg/mod
ENV GOCACHE=${GOPATH_VOLUME_DIR}/build-cache
ENV GOENV=${GOPATH_VOLUME_DIR}/env
# GOROOT MUST NOT be in the persistent volume because it may cause unintended behavior due to 
# stale toolchain files.
ENV GOROOT=/usr/local/go

ENV PATH=${GOPATH}/bin:${GOROOT}/bin:/usr/local/bin:${PATH}

WORKDIR /var/tmp/scripts
COPY tools/bootstrap/install_base_packages.sh .
COPY tools/bootstrap/install_golang.sh .
COPY tools/bootstrap/install_golang_extra_packages.sh .
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

# CI_GIT_SHA is the Git SHA of the current commit, and it only exists in GitHub Actions.
# It is used to identify the version of the application being built in BUILD_VERSION_SCRIPT.
# This argument varies across builds, so it should be placed after the validation steps to 
# maximize cache hits for the previous steps. If it is placed before the validation steps or at
# the very beginning of the Dockerfile, it will cause all subsequent steps to be re-run on every
# build, which is extremely inefficient.
ARG CI_GIT_SHA
ENV CI_GIT_SHA=${CI_GIT_SHA}

# Validate the source code by running tests and linters. This step is crucial to ensure that
# the code is in a good state before building the binary. If any of the tests or linters fail,
# the build will be stopped immediately, preventing the creation of a potentially broken binary.
RUN make fmt-check staticcheck test

# Build the application binary and move it to /app directory for the runtime image.
RUN make build \
    && mkdir -p /app \
    && mv bear-go /app

# Devcontainer image for Go development. This image may be used in both local 
# development and GitHub Actions, so additional packages for local 
# development are installed conditionally.
FROM toolchain AS dev

# XDG_VOLUME_DIR is the path of the Docker volume directory to persist XDG configurations and
# caches across container restarts. If it is not set, use the default path inside the container,
# which does not persist across restarts.
ARG XDG_VOLUME_DIR=/var/local/xdg
ENV XDG_CONFIG_HOME=${XDG_VOLUME_DIR}/config
ENV XDG_CACHE_HOME=${XDG_VOLUME_DIR}/cache
ENV XDG_DATA_HOME=${XDG_VOLUME_DIR}/data

ENV CLAUDE_CODE_EFFORT_LEVEL="high"
ENV IS_SANDBOX="1"
ENV ENABLE_LSP_TOOL="1"
ENV CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS="1"
ENV CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY="1"
ENV CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD="1"

WORKDIR /var/tmp/scripts
# Install additional packages for local development environment.
COPY tools/bootstrap/install_dev_tools.sh .
RUN chmod +x ./*.sh \
    && ./install_dev_tools.sh \
    && rm -rf /var/tmp/scripts

WORKDIR ${WORKSPACE_ROOT}

# Final runtime image to deploy the compiled binary.
FROM dhi.io/static:20250419-glibc-debian13 AS runtime

ENV BEAR_LOG_DIR=/app/logs

WORKDIR /app
COPY --from=builder /app/bear-go /app/bear-go

ENTRYPOINT ["/app/bear-go"]