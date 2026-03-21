#!/usr/bin/env bash
set -euo pipefail

# This script is executed when the `postStartCommand` action is triggered. The `postStartCommand`
# action is triggered every time the container starts. However, it is not triggered when VSCode
# opens the devcontainer and the container is already running. Changes made to the container in
# this script will persist until the container is deleted.

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi

cd "${WORKSPACE_DIR}/langgraph"
uv sync
