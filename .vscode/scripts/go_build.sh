#!/usr/bin/env bash
set -Eeuo pipefail

cd "${1:-.}"

# ProblemMatcher marker for a background process
echo "__GO_BUILD_BEGIN__"
trap 'echo "__GO_BUILD_END__"' EXIT

go build ./...