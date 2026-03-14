#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
export TZ=Asia/Seoul

npx --yes @anthropic-ai/claude-code --version
npx --yes @openai/codex --version
npx --yes @google/gemini-cli --version
