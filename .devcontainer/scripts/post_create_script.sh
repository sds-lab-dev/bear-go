#!/usr/bin/env bash
set -euo pipefail

# This script is executed when the `postCreateCommand` action is triggered. The `postCreateCommand`
# action is triggered only once right whenever the container (not the image) is newly created.
# Changes made to the container in this script will persist until the container is deleted.

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi

cd "${WORKSPACE_DIR}/langgraph"
uv venv --clear
uv sync
``