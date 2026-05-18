#!/usr/bin/env bash
set -e

REPO="git@github.com:DreamCats/byte-cli.git"
INSTALL_DIR="$HOME/.byte-cli"

echo "=== byte-cli 安装脚本 ==="

# 1. 检查/安装 Node.js
if ! command -v node &>/dev/null; then
    echo "未检测到 Node.js，正在安装..."
    if command -v brew &>/dev/null; then
        brew install node
    else
        echo "请先安装 Node.js (>=18): https://nodejs.org/"
        exit 1
    fi
else
    echo "Node.js 已安装: $(node --version)"
fi

# 2. 克隆或更新项目
if [ -d "$INSTALL_DIR" ]; then
    echo "目录已存在，拉取最新代码..."
    cd "$INSTALL_DIR" && git pull
else
    echo "克隆项目..."
    git clone "$REPO" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# 3. 安装依赖并构建
echo "安装依赖..."
npm install
echo "构建..."
npm run build

# 4. 全局链接
echo "全局链接..."
npm link

# 5. 安装 skills（给 LLM Agent 使用）
echo ""
echo "安装 skills..."
if ! command -v skills &>/dev/null; then
    echo "未检测到 skills 命令，正在安装..."
    npm install -g skills
fi

SKILLS_DIR="$INSTALL_DIR/skills"
if [ -d "$SKILLS_DIR" ]; then
    for skill in "$SKILLS_DIR"/*/; do
        skill_name=$(basename "$skill")
        echo "安装 skill: $skill_name"
        skills remove "$skill_name" -g -y 2>/dev/null || true
        skills add "$skill" -g -y 2>/dev/null || true
    done
fi

echo ""
echo "=== 安装完成 ==="
echo ""
echo "验证: byte-cli --version"
echo ""
echo "已安装的 skills:"
skills ls -g 2>/dev/null | grep "byte-" || echo "  (无)"
