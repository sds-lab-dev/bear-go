#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export TZ=Asia/Seoul

PIP=${PYTHON_VENV_DIR}/bin/pip

${PIP} install -U \
    pip \
    langgraph \
    langchain \
    langgraph-cli[inmem]
