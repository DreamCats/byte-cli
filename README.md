# byte-cli

字节内部开发工具统一 CLI（TypeScript 版）。

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/DreamCats/byte-cli/main/install.sh | bash
```

手动安装：

```bash
git clone git@github.com:DreamCats/byte-cli.git "${BYTE_CLI_HOME:-${XDG_DATA_HOME:-$HOME/.local/share}/byte-cli}"
cd "${BYTE_CLI_HOME:-${XDG_DATA_HOME:-$HOME/.local/share}/byte-cli}"
npm install && npm run build && npm link
```

更新：

```bash
cd "${BYTE_CLI_HOME:-${XDG_DATA_HOME:-$HOME/.local/share}/byte-cli}" && git pull && npm install && npm run build && npm link
```

## 开发

```bash
npm run dev -- auth status    # 直接运行源码
npm run build                  # 构建到 dist/
npm run typecheck              # 类型检查
```

## 命令一览

```
byte-cli auth login -r cn          # 认证
byte-cli auth status               # 查看认证状态
byte-cli auth config show          # 查看配置
byte-cli auth config set-cookie <cookie> -r cn

byte-cli codebase repo info <group/project>
byte-cli codebase mr get <number> -R <group/project>
byte-cli codebase mr comments <number> --unresolved

byte-cli logid <trace-id> -r us -p <psm> -k <keyword>

byte-cli psm idl <psm>
byte-cli psm api-list <psm>
byte-cli psm links <psm> -r us

byte-cli iam list -o <owner> -r cn
byte-cli iam secret <name> -r cn

byte-cli mcp list -s <psm> -r cn
byte-cli mcp tools <server-id>
byte-cli mcp call <server-id> <tool-name> -a key=value
```

## 技术栈

- TypeScript + Node.js 18+
- Commander.js — CLI 框架
- Zod — 数据校验（替代 Pydantic）
- 原生 fetch — HTTP 请求
- tsup — 构建打包

## 配置文件

```
~/.config/byte-cli/
├── config.yaml       # regions.{region}.cookie, proxy.https/http
└── token_cache/      # {region}.json
```

与 Python 版 `byte-helper` 共享同一份配置目录，可无缝切换。
