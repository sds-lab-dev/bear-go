#!/bin/bash

#
# 이 스크립트는 postCreateCommand 액션이 발생할 때 실행된다. postCreateCommand 액션은
# 컨테이너(이미지 아님)가 생성된 직후에 단 한 번만 발생한다.
#
# 이 스크립트에서 실행한 명령들이 컨테이너 내부를 변경시킨다면 해당 변경 사항은 컨테이너가 삭제될 때까지 
# 유지된다. 따라서 이 스크립트에서는 컨테이너에 추가로 설치하거나, 변경하거나, 삭제해야 하는 작업들을 
# 실행하면 된다.
#

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi

"${WORKSPACE_DIR}/.devcontainer/scripts/devcontainer-persist-links.sh" "${WORKSPACE_DIR}" || {
  echo "Failed to setup persist links" >&2
  exit 1
}