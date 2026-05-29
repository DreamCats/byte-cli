---
name: byte-mcp
description: "字节内部 MCP 平台管理。当用户需要查询 MCP Server 列表、查看工具列表、调用 MCP 工具时使用此 skill。"
---

# byte-mcp MCP 平台管理

查询 MCP Server 列表、查看工具列表、调用 MCP 工具。

## 查询 MCP Server 列表

```bash
# 查询自己的 MCP servers（默认 cn 区域）
byte-cli mcp list

# 按 PSM 搜索
byte-cli mcp list -s live_content

# 指定环境
byte-cli mcp list -e prod
byte-cli mcp list -e ppe
byte-cli mcp list -r us -e ppe

# 指定区域
byte-cli mcp list -r us
byte-cli mcp list -r eu
byte-cli mcp list -r i18n

# 分页
byte-cli mcp list -l 20 -o 10
```

## 查看 MCP Server 工具列表

```bash
# 查看指定 server 的工具列表
byte-cli mcp tools <server_id>

# 指定区域
byte-cli mcp tools <server_id> -r us
```

示例：
```bash
byte-cli mcp tools j2nkvhut
byte-cli mcp tools n8x3rz2v -r cn
```

## 调用 MCP Server 工具

```bash
# 调用工具（无参数）
byte-cli mcp call <server_id> <tool_name>

# 调用工具（带参数）
byte-cli mcp call <server_id> <tool_name> -a "key=value"

# 多个参数
byte-cli mcp call <server_id> <tool_name> -a "key1=value1" -a "key2=value2"

# JSON 值参数（array/bool/number/object 会按 JSON 解析）
byte-cli mcp call <server_id> <tool_name> \
    -a 'targets=[{"target_type":1,"product_id":123456789}]' \
    -a "dry_run=true" \
    -a "region=US"

# 指定区域
byte-cli mcp call <server_id> <tool_name> -r us
```

示例：
```bash
# 调用天气查询工具
byte-cli mcp call n8x3rz2v get_weather_d3e775f2 -a "city=北京"

# 调用用户信息工具
byte-cli mcp call j2nkvhut live_content_ai_get_user_sign_info \
    -a "oec_user_id=7494514082253145675" \
    -a "use_cache=true"
```

`mcp tools` 和 `mcp call` 只接收 `server_id` 和 `region`，不接收 `env`。如果需要测试 PPE，先用 `mcp list -e ppe` 找到对应 `server_id`，再用该 `server_id` 查询 tools 或调用工具。

## JSON 输出

```bash
byte-cli --json mcp list
byte-cli --json mcp tools <server_id>
byte-cli --json mcp call <server_id> <tool_name> -a "key=value"
byte-cli mcp call <server_id> <tool_name> -a "key=value" --json
```

## 支持区域

| 区域 | 管理 API 域名 | Server API 域名后缀 |
|------|--------------|-------------------|
| cn | cloud.bytedance.net | mcp.bytedance.net |
| us | cloud.tiktok-us.net | mcp-usttp.tiktok-us.net |
| eu | cloud-eu.tiktok-row.net | mcp-eu.tiktok-row.net |
| i18n | cloud.tiktok-row.net | mcp.tiktok-row.net |

默认区域：cn

## MCP Server 输出字段

| 字段 | 说明 |
|------|------|
| server_id | Server ID |
| psm | PSM 名称 |
| name | Server 名称 |
| description | 描述 |
| owner | 负责人 |
| admins | 管理员列表 |
| auth_enabled | 是否启用认证 |
| current_revision_id | 当前版本 ID |
| created_at | 创建时间 |
| updated_at | 更新时间 |

## 工具输出字段

| 字段 | 说明 |
|------|------|
| name | 工具名称 |
| description | 工具描述 |
| inputSchema.type | 参数类型 |
| inputSchema.properties | 参数定义 |
| inputSchema.required | 必填参数列表 |

## 使用场景

- 查看自己或团队的 MCP Server 列表
- 查看某个 MCP Server 提供的工具
- 调用 MCP 工具进行测试或调试
- 集成到自动化脚本中调用 MCP 服务
