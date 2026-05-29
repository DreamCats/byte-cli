---
name: byte-repo
description: "字节内部代码仓库和 MR 查询。当用户需要查看仓库信息、MR 详情、MR 评论、Code Review 时使用此 skill。"
---

# byte-repo 仓库查询

查询字节内部代码仓库和 MR 信息。

## 查询仓库信息

```bash
byte-cli codebase repo info <group/project>
```

## 查看 MR 详情

```bash
# 自动推断仓库（需在 git 仓库中）
byte-cli codebase mr get <number>

# 指定仓库
byte-cli codebase mr get <number> --repo <group/project>
byte-cli codebase mr get <number> -R <group/project>
```

## 查看 MR 评论

```bash
# 查看所有评论
byte-cli codebase mr comments <number>
byte-cli codebase mr comments <number> -R <group/project>

# 只看未解决评论
byte-cli codebase mr comments <number> --unresolved
byte-cli codebase mr comments <number> -R <group/project> --unresolved
```

## JSON 输出

```bash
byte-cli --json codebase repo info <group/project>
byte-cli --json codebase mr get <number> -R <group/project>
byte-cli --json codebase mr comments <number> -R <group/project>
byte-cli codebase mr comments <number> -R <group/project> --json
```

## 自动推断仓库

在 git 仓库中运行时，可从 remote URL 自动推断仓库名，无需手动指定。
不在目标 git 仓库中，或 remote 无法推断 Codebase 仓库名时，需要使用 `-R/--repo <group/project>`。

## 评论解析

aime 机器人评论会自动解析：
- 问题描述
- 优先级（P0 红色、P1 黄色）
- 问题分类
