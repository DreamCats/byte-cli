---
name: byte-log
description: "字节内部分布式日志链路查询。当用户需要通过 Log ID 查询日志、排查调用链路问题、查看服务间调用日志时使用此 skill。"
---

# byte-log 日志查询

通过 Log ID 查询分布式日志链路。

## 基本查询

```bash
byte-cli logid <trace-id> --region us
byte-cli logid <trace-id> --region i18n
byte-cli logid <trace-id> --region eu
```

注意：cn 区域暂不支持日志查询。

## PSM 过滤

```bash
byte-cli logid <trace-id> --region us --psm service1 --psm service2
byte-cli logid <trace-id> -r i18n -p ttec.industry.solution
```

## 关键词过滤

```bash
byte-cli logid <trace-id> --region us --keyword error --keyword timeout
byte-cli logid <trace-id> -r us -k error -k timeout
```

关键词为 OR 关系，不区分大小写。

## 日志级别过滤

```bash
byte-cli logid <trace-id> -r us --level error --level warn
byte-cli logid <trace-id> -r us -l error -l warn
```

## 消息截断

```bash
byte-cli logid <trace-id> --region us --max-len 500
byte-cli logid <trace-id> -r us --max-len 0
```

默认 300 字符，0 表示不截断。

## 条数限制

```bash
byte-cli logid <trace-id> -r us --limit 50
```

普通输出默认最多显示 20 条。

## JSON 输出

```bash
byte-cli --json logid <trace-id> --region us
byte-cli logid <trace-id> -r us --json
```

JSON 输出不限制条数和消息长度。

## 支持区域

| 区域 | 日志服务地址 |
|------|-------------|
| us | logservice-tx.tiktok-us.org |
| i18n | logservice-sg.tiktok-row.org |
| eu | logservice-eu-ttp.tiktok-eu.org |
