#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export TZ=Asia/Seoul
export GOROOT=$HOME/.go
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/bin:$HOME/.local/bin:$GOPATH/bin:$GOROOT/bin

go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest