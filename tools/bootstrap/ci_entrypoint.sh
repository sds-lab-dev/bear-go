#!/usr/bin/env bash
set -euo pipefail

# This script is the entrypoint for the GitHub Actions CI container.

source /usr/local/bin/start_dockerd.sh
source /usr/local/bin/install_langgraph_dependency.sh

main() {
  start_dockerd
  install_langgraph_dependency

  # Execute the command passed as arguments to the entrypoint. This allows the container to run the
  # default command specified in the Dockerfile or any command from the devcontainer. If this step
  # is omitted, the container will not execute the default command and will exit immediately after
  # running the entrypoint script. 
  exec "$@"
}

main "$@"
