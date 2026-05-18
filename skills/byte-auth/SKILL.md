---
name: byte-auth
description: "字节内部多区域统一认证管理。当用户需要配置 Cookie、获取 Token、查看认证状态、切换区域时使用此 skill。"
---

# byte-auth 认证管理

管理字节内部各区域的认证信息。

## 查看认证状态

```bash
byte-cli auth status
```

显示各区域 Cookie 和 Token 配置状态。

## 配置 Cookie

```bash
byte-cli auth config set-cookie <value> --region cn
byte-cli auth config set-cookie <value> --region us
byte-cli auth config set-cookie <value> --region codebase
```

## 获取 Token

```bash
byte-cli auth token --region cn
byte-cli auth token --region us
byte-cli auth token --region codebase
```

Token 会自动缓存，1 小时有效期。

## 查看配置

```bash
byte-cli auth config show
```

## 区域说明

| 区域 | Cookie 名 | 用途 |
|------|-----------|------|
| cn | CAS_SESSION | 国内服务 |
| i18n | CAS_SESSION | 国际化服务 |
| us | CAS_SESSION | 美国服务 |
| eu | CAS_SESSION | 欧洲服务 |
| codebase | CAS_SESSION_API | 代码仓库服务 |
