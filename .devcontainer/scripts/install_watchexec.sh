#!/usr/bin/env bash
set -euo pipefail

VERSION="2.4.1"
ARCH="$(uname -m)"

case "${ARCH}" in
    x86_64|amd64)
        ARCH_SUFFIX="x86_64-unknown-linux-gnu"
        ;;
    aarch64|arm64)
        ARCH_SUFFIX="aarch64-unknown-linux-gnu"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}" >&2
        exit 1
        ;;
esac

URL="https://github.com/watchexec/watchexec/releases/download/v${VERSION}/watchexec-${VERSION}-${ARCH_SUFFIX}.tar.xz"

echo "Detected arch: ${ARCH}"
echo "Archecture suffix: ${ARCH_SUFFIX}"
echo "Downloading: ${URL}"

wget -q --show-progress -O "/tmp/watchexec.tar.xz" -L "${URL}"
mkdir -p /usr/local/bin
tar -xJf /tmp/watchexec.tar.xz \
  -C /usr/local/bin \
  --strip-components=1 \
  watchexec-${VERSION}-${ARCH_SUFFIX}/watchexec
chmod +x /usr/local/bin/watchexec