#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export TZ=Asia/Seoul

go install -v github.com/go-delve/delve/cmd/dlv@latest