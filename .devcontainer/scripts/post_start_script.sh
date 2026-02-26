#!/bin/bash

#
# 이 스크립트는 postStartCommand 액션이 발생할 때 실행된다. postStartCommand 액션은 컨테이너가
# 실행될 때마다 매번 발생한다.
#
# 따라서 이 스크립트에서는 컨테이너가 매번 실행될 때마다 미리 실행주어야 하는 작업들을 실행하면 된다.
#

WORKSPACE_DIR="${1:-}"
if [[ -z "$WORKSPACE_DIR" ]]; then
  echo "Workspace path argument missing" >&2
  exit 1
fi

git config --global credential.helper "store --file /root/.persist/git-credentials"
git config --global rerere.enabled true
git config --global rerere.autoupdate true
git config --global alias.graph "log --graph --oneline"
git config --global alias.full "log --pretty=fuller"
git config --global alias.pr '!f() { base=$(git merge-base ${1:-main} HEAD) && git diff "$base"...HEAD; }; f'
git config --global alias.pr-wip '!f() { base=$(git merge-base ${1:-main} HEAD) && git diff "$base"; }; f'
git config --global alias.rebase-pr '!f() { base=$(git merge-base --fork-point ${1} ${2} 2>/dev/null || git merge-base ${1} ${2}) && git rebase --onto main "$base" ${2}; }; f'
git config --global alias.merge-pr "merge --no-ff --no-commit"
git config --global alias.cleanup "!git fetch --prune && git branch -vv | grep ': gone]' | awk '{print \$1}' | xargs -r git branch -d"
git config --global core.checkStat minimal
git config --global core.trustctime false
git config --global core.fsmonitor false
git config --global core.filemode false
git config --global merge.conflictstyle zdiff3
git config --global gc.reflogExpire 360.days
git config --global gc.reflogExpireUnreachable 180.days
git config --global push.autoSetupRemote true
git config --global pull.ff only
