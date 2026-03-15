#!/usr/bin/env bash
set -euo pipefail

# This script is the entrypoint for the development container. It is executed every time the
# container starts with root permissions.
# 
# So, it can be used as an alternative to using the `postCreateCommand` or `postStartCommand`
# hooks in `devcontainer.json` that are only executed once after the container is newly created or
# started (only from stopped containers), respectively. 
#
# If you need to run some logic whenever the devcontainer project starts regardless of the
# container is already started or not, you can add it to `.devcontainer/bashrc-settings`.

# This git config settings use the `--system` flag to ensure that the devcontainer server copies
# the hosts's git configurations. If we use the `--global` flag instead, the git configurations of
# the host will be ignored by the devcontainer server.
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

  # Execute the command passed as arguments to the entrypoint. This allows the container to run the
  # default command specified in the Dockerfile or any command from the devcontainer. If this step
  # is omitted, the container will not execute the default command and will exit immediately after
  # running the entrypoint script. 
  exec "$@"
}

main "$@"
