#!/usr/bin/env bash
set -euo pipefail

#
# 이 스크립트는 initializeCommand 액션이 발생할 때 실행된다. initializeCommand 액션은 이미지를 
# 생성하기 전에 먼저 실행된다. 따라서 이 스크립트에서는 이미지 생성 전에 미리 호스트에서 준비해야 할 
# 작업들을 실행하면 된다.
#

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi
