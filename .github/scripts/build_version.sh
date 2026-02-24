#!/usr/bin/env bash
set -euo pipefail

REV_LEN="${REV_LEN:-12}"
DIFF_LEN="${DIFF_LEN:-12}"
DIFF_PAD_CHAR="${DIFF_PAD_CHAR:-0}"

# Get the short git revision hash if we're in a git repository.
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    GIT_REV="$(git rev-parse --short="${REV_LEN}" HEAD 2>/dev/null || true)"
else
    GIT_REV=""
fi

if [[ -z "${GIT_REV}" ]]; then
  echo "ERROR: git rev-parse failed." >&2
  exit 1
fi

if [[ -z "$(git status --porcelain=v1 --untracked-files=normal 2>/dev/null || true)" ]]; then
    # If there are no uncommitted/untracked changes, use a string of zeros for 
    # the diff ID.
    DIFF_ID="$(printf "%*s" "${DIFF_LEN}" "" | tr ' ' "${DIFF_PAD_CHAR}")"
else
    # Othwerise, compute a hash of the workspace state. This is done by hashing 
    # the contents of all tracked and untracked files. The resulting hash is then
    # truncated to the specified length to form the diff ID.
    WORKSPACE_HASH="$(
        tmp="$(mktemp)"
        trap 'rm -f "$tmp"' RETURN

        while IFS= read -r -d '' f; do
        if [[ -f "$f" ]]; then
            h="$(sha256sum -- "$f" | awk '{print $1}')"
            printf '%s\t%s\n' "$f" "$h" >>"$tmp"
        fi
        done < <(git ls-files -z --cached --others --exclude-standard 2>/dev/null)

        # Sort the file hashes to ensure consistent ordering, then hash the 
        # combined output.
        sort "$tmp" | sha256sum | awk '{print $1}'
    )"
    DIFF_ID="${WORKSPACE_HASH:0:${DIFF_LEN}}"
fi

# VERSION is the combination of the git revision and the diff ID that represents 
# the current state of the workspace, allowing us to distinguish between different
# states even if they share the same git revision.
VERSION="${GIT_REV}-${DIFF_ID}"
printf '%s\n' "${VERSION}"