#!/usr/bin/env bash
set -euo pipefail

# This script is the entrypoint for the development container. It is executed every time the
# container starts with root permissions.
# 
# So, it can be used as an alternative to using the `postCreateCommand` or `postStartCommand`
# hooks in `devcontainer.json` that are only executed once after the container is newly created or
# started (only from stopped containers), respectively. 
#
# If you need to run some logic whenever the container starts regardless of its previous state,
# you can add it to `.devcontainer/bashrc-settings`. However, this file is sourced with non-root
# permissions.

set_git_configs() {
  git config --global credential.helper "store --file $GIT_CREDENTIALS_DIR/git-credentials"
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
  git config --global pull.rebase true  
  git config --global core.hooksPath .githooks
}

main() {
  set_git_configs

  # Execute the command passed as arguments to the entrypoint. This allows the container to run the
  # default command specified in the Dockerfile or any command from the devcontainer. If this step
  # is omitted, the container will not execute the default command and will exit immediately after
  # running the entrypoint script. 
  exec "$@"
}

main "$@"
