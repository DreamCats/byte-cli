#!/usr/bin/env bash
set -euo pipefail

MODULE="github.com/DreamCats/byte-cli/cmd/byte-cli@latest"
LINK_DIR="$HOME/.local/bin"

echo "=== byte-cli 安装脚本 ==="

if ! command -v go >/dev/null 2>&1; then
  echo "未检测到 Go，请先安装 Go 1.22+：https://go.dev/dl/"
  exit 1
fi

echo "Go: $(go version)"
echo "安装 byte-cli..."
GOBIN="$LINK_DIR" go install "$MODULE"
echo "已安装: $LINK_DIR/byte-cli"

if [[ ":$PATH:" != *":$LINK_DIR:"* ]]; then
  echo "提示: 请将 $LINK_DIR 添加到 PATH"
  echo "  export PATH=\"$LINK_DIR:\$PATH\""
fi

if command -v skills >/dev/null 2>&1; then
  REPO_DIR="${BYTE_CLI_HOME:-}"
  if [[ -z "$REPO_DIR" ]]; then
    REPO_DIR="$(pwd)"
  fi
  SKILLS_DIR="$REPO_DIR/skills"
  if [[ -d "$SKILLS_DIR" ]]; then
    echo ""
    echo "安装 skills..."
    for skill in "$SKILLS_DIR"/byte-*; do
      [[ -d "$skill" ]] || continue
      skill_name="$(basename "$skill")"
      echo "安装 skill: $skill_name"
      skills remove "$skill_name" -g -y 2>/dev/null || true
      skills add "$skill" -g -y 2>/dev/null || true
    done
  fi
else
  echo "未检测到 skills 命令，跳过 skills 安装"
fi

echo ""
echo "=== 安装完成 ==="
echo "验证: byte-cli --version"
