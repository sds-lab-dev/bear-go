#!/usr/bin/env bash
set -euo pipefail

PERSIST="${PERSIST:-/persist}"

# Absolute paths only
DIR_TARGETS=(
  "/root/.claude"
  "/root/.local"
  "/root/go"
  "/root/.cache/go-build"
  "/root/.config/go"
)

FILE_TARGETS=(
  "/workspace/.env"
  "/root/.git-credentials"
)

ts() { date +%Y%m%d-%H%M%S; }

sha256_str() {
  local s="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s' "$s" | sha256sum | awk '{print $1}'
  elif command -v openssl >/dev/null 2>&1; then
    printf '%s' "$s" | openssl dgst -sha256 | awk '{print $NF}'
  else
    echo "Need sha256sum or openssl" >&2
    exit 1
  fi
}

# type: "d" (dir) or "f" (file)
ensure_persist_link() {
  local type="$1"
  local target="$2"
  [[ "$target" == /* ]] || { echo "Target must be absolute: $target" >&2; exit 1; }

  local h dst
  h="$(sha256_str "$target")"
  dst="${PERSIST}/${h}"

  # If already a symlink, fix only if wrong.
  if [[ -L "$target" ]]; then
    local cur
    cur="$(readlink -- "$target" || true)"
    if [[ "$cur" != "$dst" ]]; then
      rm -f -- "$target"
      ln -s -- "$dst" "$target"
    fi
    return 0
  fi

  # If target exists as a real path:
  # - If persist dst doesn't exist yet, move target into persist (first-run 
  #   migration).
  # - Else, keep persist as source of truth and back up the target.
  if [[ -e "$target" ]]; then
    if [[ ! -e "$dst" ]]; then
      mkdir -p -- "$(dirname -- "$dst")"
      mv -- "$target" "$dst"
    else
      mv -- "$target" "${target}.bak.$(ts)"
    fi
  fi

  # If persist dst still doesn't exist (target didn't exist), create a placeholder.
  if [[ ! -e "$dst" ]]; then
    if [[ "$type" == "d" ]]; then
      mkdir -p -- "$dst"
    else
      mkdir -p -- "$(dirname -- "$dst")"
      : > "$dst"
    fi
  fi

  # Ensure parent dir exists, then link.
  mkdir -p -- "$(dirname -- "$target")"
  ln -s -- "$dst" "$target"
}

main() {
  mkdir -p -- "$PERSIST"

  for d in "${DIR_TARGETS[@]}"; do
    ensure_persist_link "d" "$d"
  done

  for f in "${FILE_TARGETS[@]}"; do
    ensure_persist_link "f" "$f"
  done

  # Minimal SSH permission hardening (optional but practical)
  local ssh_dst="${PERSIST}/$(sha256_str "/root/.ssh")"
  if [[ -d "$ssh_dst" ]]; then
    chmod 700 -- "$ssh_dst"
    find -- "$ssh_dst" -maxdepth 1 -type f -name 'id_*' ! -name '*.pub' -exec chmod 600 {} + 2>/dev/null || true
  fi
}

main "$@"