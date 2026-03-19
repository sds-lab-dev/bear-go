#!/usr/bin/env bash

install_langgraph_dependency() {
  local PIP="${PYTHON_VENV_DIR}/bin/pip"

  ${PIP} install -e "${DEVCONTAINER_WORKSPACE}/langgraph"
}