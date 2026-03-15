#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export TZ=Asia/Seoul

curl -fsSL https://claude.ai/install.sh | bash
npx --yes @openai/codex --version
npx --yes @google/gemini-cli --version
