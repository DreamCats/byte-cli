---
name: byte-psm
description: "字节内部 PSM 服务信息查询。当用户需要查询某个 PSM 的 IDL 仓库路径、API 接口列表、平台链接时使用此 skill。"
---

# byte-psm PSM 信息查询

查询 PSM 对应的 IDL 信息、API 接口列表、平台链接。

## 查询 IDL 信息

```bash
byte-cli psm idl <psm-name>
```

示例：
```bash
byte-cli psm idl ttec.content.product_model
```

## 查询 API 接口列表

```bash
byte-cli psm api-list <psm-name>
```

示例：
```bash
byte-cli psm api-list ttec.content.product_model
```

## 生成平台链接

```bash
byte-cli psm links <psm-name> --region us
byte-cli psm links <psm-name> -r eu
byte-cli psm links <psm-name> -r i18n
```

支持区域：us、eu、i18n（默认 us）

## JSON 输出

```bash
byte-cli --json psm idl <psm-name>
byte-cli --json psm api-list <psm-name>
byte-cli --json psm links <psm-name> -r us
```

## IDL 输出字段

| 字段 | 说明 |
|------|------|
| PSM | 服务标识 |
| RepoName | IDL 仓库名（如 oec/rpc_idl） |
| IDLPath | IDL 文件路径 |
| DefaultBranch | 默认分支 |
| IDLVersion | IDL 版本号 |
| IDLRepoURL | IDL 仓库完整 URL |
| IDLRepoMainIDLURL | IDL 主文件完整 URL |

## API 接口输出字段

| 字段 | 说明 |
|------|------|
| name | 接口方法名 |
| note | 接口说明 |
| owner | 负责人 |
| version | 版本号 |
| serializer | 序列化方式（json/thrift） |
| oneway | 是否单向调用 |
| updated_at | 最后更新时间 |

## 平台链接说明

| 平台 | 说明 |
|------|------|
| TCE | 服务管理平台 |
| SCM | 编译产物平台 |
| TCC | 配置中心 |
| Overpass | Overpass 平台（cn 区域） |

## 使用场景

- 快速定位某个服务的 IDL 定义位置
- 查看服务提供了哪些 API 接口
- 获取接口负责人信息
- 快速跳转到各平台查看服务详情
