---
name: byte-iam
description: "字节内部 IAM 服务账号管理。当用户需要查询服务账号列表、获取服务账号密钥时使用此 skill。"
---

# byte-iam 服务账号管理

查询 IAM 服务账号列表和密钥。

## 查询服务账号列表

```bash
# 查询自己的服务账号（默认 cn 区域）
byte-cli iam list

# 查询指定用户的服务账号
byte-cli iam list -o <username>

# 指定区域
byte-cli iam list -r us
byte-cli iam list -r eu
byte-cli iam list -r i18n

# 分页
byte-cli iam list -p 2 -s 20
```

## 查询服务账号密钥

此命令会打印服务账号密钥，只在用户明确需要查看密钥时运行，避免在普通排查中调用。

```bash
byte-cli iam secret <account-name>
byte-cli iam secret <account-name> -r us
```

示例：
```bash
byte-cli iam secret ecom.page.tag
byte-cli iam secret ttec_script_monitor -r us
```

## JSON 输出

```bash
byte-cli --json iam list
byte-cli --json iam list -o <username>
byte-cli --json iam secret <account-name>
```

## 支持区域

| 区域 | 域名 |
|------|------|
| cn | cloud.bytedance.net |
| us | cloud.tiktok-us.net |
| eu | cloud-eu.tiktok-row.net |
| i18n | cloud.tiktok-row.net |

默认区域：cn

## 服务账号输出字段

| 字段 | 说明 |
|------|------|
| name | 服务账号名称 |
| id | 账号 ID |
| spec.name | 完整名称 |
| spec.path | 组织路径 |
| spec.description | 描述 |
| spec.owners | 负责人列表 |
| spec.created_by | 创建人 |
| spec.created_at | 创建时间 |
| spec.secret | 密钥（secret 命令） |
| spec.secrets | 密钥列表（含 uid、enabled 状态） |

## 使用场景

- 查看自己或他人的服务账号列表
- 获取服务账号的密钥 key
- 快速定位服务账号的组织路径和负责人
