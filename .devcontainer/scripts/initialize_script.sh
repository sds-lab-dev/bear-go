#!/usr/bin/env bash
set -euo pipefail

# This script is executed when the `initializeCommand` action is triggered. The `initializeCommand`
# action is executed on the host (not inside the container) before the image is newly created or
# subsequently started. Note that this script is executed every time the devcontainer is opened.

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi
