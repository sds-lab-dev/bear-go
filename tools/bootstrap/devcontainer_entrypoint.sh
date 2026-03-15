#!/usr/bin/env bash
set -euo pipefail

# This script is the entrypoint for the development container. It is executed every time the
# container starts with non-root permissions.
# 
# So, it can be used as an alternative to using the `postCreateCommand` or `postStartCommand`
# hooks in `devcontainer.json` that are only executed once after the container is newly created or
# started (only from stopped containers), respectively. 
#
# If you need to run some logic whenever the devcontainer project starts regardless of the
# container is already started or not, you can add it to `.devcontainer/bashrc-settings`.

main() {
  # Execute some logics here that you need to run whenever the devcontainer project starts.

  # Execute the command passed as arguments to the entrypoint. This allows the container to run the
  # default command specified in the Dockerfile or any command from the devcontainer. If this step
  # is omitted, the container will not execute the default command and will exit immediately after
  # running the entrypoint script. 
  exec "$@"
}

main "$@"
