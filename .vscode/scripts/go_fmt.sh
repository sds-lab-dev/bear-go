#!/usr/bin/env bash
set -Eeuo pipefail

cd "${1:-.}"

# ProblemMatcher marker for a background process
echo "__GO_FMT_BEGIN__"
trap 'echo "__GO_FMT_END__"' EXIT

# Only check .go files tracked by git to avoid unnecessary formatting
FILES="$(git ls-files '*.go')"

# List unformatted files, then format them in place if any are found
UNFORMATTED="$(echo $FILES | xargs -r gofmt -l)"

if [[ -n "$UNFORMATTED" ]]; then
  # Process the list of unformatted files safely line by line
  printf '%s\n' "$UNFORMATTED" | xargs -r gofmt -w
fi