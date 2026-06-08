---
name: byte-metrics
description: "字节内部 Metrics 指标查询。当用户需要查询 metric 数据、tag keys 或 fields 时使用此 skill。"
---

# byte-metrics Metrics 指标查询

查询 metric 数据、tag keys 和 fields。

## 查询 Tag Keys

```bash
byte-cli metrics tagk <metric_name>
byte-cli metrics tagk <metric_name> -r us
```

已选 tag 过滤条件可以用 `--tagkv key=value` 传入，可重复：

```bash
byte-cli metrics tagk <metric_name> --tagkv result=hit --tagkv scope=global
```

## 查询指标数据

```bash
byte-cli metrics query <metric_name>
byte-cli metrics query <metric_name> -r us
```

默认查询最近 30 分钟，不带 group key。需要按 tag key 分组时使用
`--group-by`，可重复：

```bash
byte-cli metrics query <metric_name> --group-by topic
```

tag 过滤条件使用 `--filter key=value`，可重复。

## 查询 Fields

```bash
byte-cli metrics field <metric_name>
byte-cli metrics field <metric_name> -r us
```

## JSON 输出

```bash
byte-cli --json metrics tagk <metric_name>
byte-cli --json metrics query <metric_name>
byte-cli --json metrics field <metric_name>
```

## 支持区域

当前内置区域：

- `us`：对应平台参数 `_region=US-TTP`
- `i18n`：对应平台参数 `_region=Singapore-Central`

未内置的区域可以临时指定平台地址：

```bash
byte-cli metrics tagk <metric_name> \
  -r eu \
  --metrics-region EU-TTP \
  --endpoint https://metrics.example/api \
  --origin https://metrics-fe.example
```

## 注意

Metrics 平台使用 `Authorization` 头承载当前区域的 JWT。命令会复用
`byte-cli auth token -r <region>` 的缓存和刷新逻辑。
