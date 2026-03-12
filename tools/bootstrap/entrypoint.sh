#!/usr/bin/env bash
set -euo pipefail

# This script is the entrypoint for the development container. It is executed every time the
# container starts, so it is alternative to using the `postCreateCommand` or `postStartCommand`
# hooks in `devcontainer.json` that reveal annoying terminal output.

# REMOTE_USER MUST be the same as `remoteUser` and `username` of common-utils feature specified in
# `devcontainer.json`.
REMOTE_USER="devuser"

# Set ownership of the named volumes to the non-root user to ensure they are writable.
sudo chown -R "$REMOTE_USER:$REMOTE_USER" \
  "$GOPATH" \
  "$GOROOT" \
  "$GIT_CREDENTIALS_DIR" \
  "$XDG_CONFIG_HOME" \
  "$XDG_CACHE_HOME" \
  "$XDG_DATA_HOME" \
  "$BEAR_LOG_DIR"

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
git config --global pull.ff only
git config --global core.hooksPath .githooks

# Execute the command passed as arguments to the entrypoint. This allows the container to run the
# default command specified in the Dockerfile or any command from the devcontainer. 
exec "$@"