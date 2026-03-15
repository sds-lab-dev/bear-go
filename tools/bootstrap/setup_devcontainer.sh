#!/usr/bin/env bash
set -euo pipefail

# This script is to set up the development container with root permissions. It is executed only
# once when the container is newly created.

set_git_configs() {
  git config --system rerere.enabled true
  git config --system rerere.autoupdate true
  git config --system alias.graph "log --graph --oneline"
  git config --system alias.full "log --pretty=fuller"
  git config --system alias.pr '!f() { base=$(git merge-base ${1:-main} HEAD) && git diff "$base"...HEAD; }; f'
  git config --system alias.pr-wip '!f() { base=$(git merge-base ${1:-main} HEAD) && git diff "$base"; }; f'
  git config --system alias.rebase-pr '!f() { base=$(git merge-base --fork-point ${1} ${2} 2>/dev/null || git merge-base ${1} ${2}) && git rebase --onto main "$base" ${2}; }; f'
  git config --system alias.merge-pr "merge --no-ff --no-commit"
  git config --system alias.cleanup "!git fetch --prune && git branch -vv | grep ': gone]' | awk '{print \$1}' | xargs -r git branch -d"
  git config --system core.checkStat minimal
  git config --system core.trustctime false
  git config --system core.fsmonitor false
  git config --system core.filemode false
  git config --system merge.conflictstyle zdiff3
  git config --system gc.reflogExpire 360.days
  git config --system gc.reflogExpireUnreachable 180.days
  git config --system push.autoSetupRemote true
  git config --system pull.rebase true  
  git config --system core.hooksPath .githooks
}

main() {
  set_git_configs
}

main "$@"