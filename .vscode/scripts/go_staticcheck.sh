#!/usr/bin/env bash
set -Eeuo pipefail

cd "${1:-.}"

# ProblemMatcher marker for a background process
echo "__GO_STATICCHECK_BEGIN__"
trap 'echo "__GO_STATICCHECK_END__"' EXIT

staticcheck ./...