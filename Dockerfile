# Toolchain image to build Go.
FROM debian:13 AS toolchain

SHELL ["/bin/bash", "-c"]

ENV TZ=Asia/Seoul

# OVERLAYS_DIR is the base path to group all directories that are used as overlay volumes in this
# image. It does not persist across container restarts. Each subdirectory under it is used for a
# specific purpose, such as storing Rust toolchain and Cargo cache files.
ARG OVERLAYS_DIR=/var/overlays

# GOPATH_DIR is the path of the Docker volume to persist Go compiled artifacts to speed up
# subsequent builds. If it is not set, use the default path inside the container, which does not
# persist across restarts.
ARG GOPATH_DIR=${OVERLAYS_DIR}/go
ENV GOPATH=${GOPATH_DIR}
ARG GOMODCACHE_DIR=${GOPATH_DIR}/pkg/mod
ENV GOMODCACHE=${GOMODCACHE_DIR}
ARG GOCACHE_DIR=${GOPATH_DIR}/build-cache
ENV GOCACHE=${GOCACHE_DIR}
ENV GOENV=${GOPATH}/env
# GOROOT_DIR MUST NOT be in the persistent volume because it may cause unintended behavior due to 
# stale toolchain files.
ARG GOROOT_DIR=${OVERLAYS_DIR}/go-root
ENV GOROOT=${GOROOT_DIR}

ARG NPM_CONFIG_CACHE_DIR=${OVERLAYS_DIR}/npm
# npm_config_cache MUST be lowercase environment variable.
ENV npm_config_cache=${NPM_CONFIG_CACHE_DIR}

ENV PATH=${GOPATH}/bin:${GOROOT}/bin:/usr/local/bin:${PATH}

WORKDIR /var/tmp/scripts
COPY tools/bootstrap/install_base_packages.sh .
COPY tools/bootstrap/install_golang.sh .
COPY tools/bootstrap/install_golang_extra_packages.sh .
RUN chmod +x ./*.sh && \
    ./install_base_packages.sh && \
    ./install_golang.sh --version 1.26.0 && \
    ./install_golang_extra_packages.sh && \
    rm -rf /var/tmp/scripts

ENV LANG=en_US.UTF-8
ENV LC_CTYPE=ko_KR.UTF-8
ENV LESSCHARSET=utf-8

COPY .devcontainer/bashrc-settings /tmp/bashrc-settings
RUN cat /tmp/bashrc-settings >> /etc/bash.bashrc && \
    printf "\n" >> /etc/bash.bashrc && \
    rm /tmp/bashrc-settings && \
    localedef -f UTF-8 -i ko_KR ko_KR.UTF-8 && \
    localedef -f UTF-8 -i en_US en_US.UTF-8

WORKDIR /opt/devcontainer

# Builder image to pass a compiled binary to the final runtime image.
FROM toolchain AS builder

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
RUN --mount=type=cache,id=bear-go-mod-cache,target=${GOMODCACHE} \
    --mount=type=cache,id=bear-go-build-cache,target=${GOCACHE} \
    make fmt-check staticcheck test

# Build the application binary and move it to /app directory for the runtime image.
RUN --mount=type=cache,id=bear-go-mod-cache,target=${GOMODCACHE} \
    --mount=type=cache,id=bear-go-build-cache,target=${GOCACHE} \
    make build && \
    mkdir -p /app && \
    mv bear-go /app

# Devcontainer image for Go development. This image may be used in both local 
# development and GitHub Actions, so additional packages for local 
# development are installed conditionally.
FROM toolchain AS dev

# GIT_CREDENTIALS_DIR is the path of the Docker volume directory to persist Git credentials across
# container restarts. If it is not set, use the default path inside the container, which does not
# persist across restarts.
ARG GIT_CREDENTIALS_DIR=${OVERLAYS_DIR}/git-credentials
ENV GIT_CREDENTIALS_DIR=${GIT_CREDENTIALS_DIR}

# XDG_DIR is the path of the Docker volume directory to persist XDG configurations and caches
# across container restarts. If it is not set, use the default path inside the container, which
# does not persist across restarts.
ARG XDG_DIR=${OVERLAYS_DIR}/xdg
ENV XDG_CONFIG_HOME=${XDG_DIR}/config
ENV XDG_CACHE_HOME=${XDG_DIR}/cache
ENV XDG_DATA_HOME=${XDG_DIR}/data

ARG CLAUDE_CONFIG_DIR=${OVERLAYS_DIR}/claude
ENV CLAUDE_CONFIG_DIR=${CLAUDE_CONFIG_DIR}
ENV CLAUDE_CODE_EFFORT_LEVEL="high"
ENV IS_SANDBOX="1"
ENV ENABLE_LSP_TOOL="1"
ENV CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS="1"
ENV CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY="1"
ENV CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD="1"
ENV DISABLE_AUTOUPDATER="1"

ARG CODEX_HOME=${OVERLAYS_DIR}/codex
ENV CODEX_HOME=${CODEX_HOME}

ARG GEMINI_CLI_HOME=${OVERLAYS_DIR}/gemini
ENV GEMINI_CLI_HOME=${GEMINI_CLI_HOME}

WORKDIR /var/tmp/scripts
# Install additional packages for local development environment.
COPY tools/bootstrap/install_ai_assistants.sh .
RUN chmod +x ./*.sh && \
    mkdir -p \
    "${GIT_CREDENTIALS_DIR}" \
    "${XDG_CONFIG_HOME}" \
    "${XDG_CACHE_HOME}" \
    "${XDG_DATA_HOME}" \
    "${CLAUDE_CONFIG_DIR}" \
    "${CODEX_HOME}" \
    "${GEMINI_CLI_HOME}" && \
    ./install_ai_assistants.sh && \
    rm -rf /var/tmp/scripts

COPY tools/bootstrap/devcontainer_entrypoint.sh /usr/local/bin/devcontainer_entrypoint.sh
RUN chmod +x /usr/local/bin/devcontainer_entrypoint.sh

WORKDIR /opt/devcontainer

ENTRYPOINT ["/usr/local/bin/devcontainer_entrypoint.sh"]
CMD [ "sleep", "infinity" ]

# Final runtime image to deploy the compiled binary.
FROM dhi.io/static:20250419-glibc-debian13 AS runtime

ENV BEAR_LOG_DIR=/app/logs

ARG APP_UID=1001
ARG APP_GID=1001

WORKDIR /app
COPY --from=builder --chown=${APP_UID}:${APP_GID} /app/bear-go /app/bear-go

USER ${APP_UID}:${APP_GID}
ENTRYPOINT ["/app/bear-go"]
