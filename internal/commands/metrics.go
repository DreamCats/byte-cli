package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DreamCats/byte-cli/internal/httpclient"
)

const defaultMetricsTenant = "default"
const defaultMetricsQueryWindow = "30m"

var metricsRegionMap = map[string]metricsRegionConfig{
	"us": {
		AuthRegion:     "us",
		PlatformRegion: "US-TTP",
		APIBase:        "https://metrics-svc-platform-ttp.tiktok-us.org/byteplot/api/v2",
		QueryAPIBase:   "https://metrics-svc-platform-ttp.tiktok-us.org/byteplot/api",
		Origin:         "https://metrics-fe-ttp-us.tiktok-row.org",
		FetchSite:      "cross-site",
	},
	"i18n": {
		AuthRegion:     "i18n",
		PlatformRegion: "Singapore-Central",
		APIBase:        "https://metrics-fe-i18n.tiktok-row.org/byteplot/api",
		Origin:         "https://metrics-fe-i18n.tiktok-row.org",
		Referer:        "https://metrics-fe-i18n.tiktok-row.org/web/plot/metrics",
		FetchSite:      "same-origin",
	},
}

type metricsRegionConfig struct {
	AuthRegion     string
	PlatformRegion string
	APIBase        string
	QueryAPIBase   string
	Origin         string
	Referer        string
	FetchSite      string
}

type metricsOptions struct {
	Region        string
	MetricsRegion string
	Tenant        string
	APIBase       string
	Origin        string
	TagKVs        []metricsTagKV
}

type metricsQueryOptions struct {
	Metrics    metricsOptions
	Aggregator string
	Field      string
	TopK       string
	Downsample string
	Start      string
	End        string
	Window     string
	GroupBys   []string
	Filters    []metricsTagKV
}

type metricsTagKV struct {
	Key   string
	Value string
}

type MetricsSuggestResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

type MetricsQueryOutput struct {
	Metric         string `json:"metric"`
	Region         string `json:"region"`
	PlatformRegion string `json:"platform_region"`
	Start          int64  `json:"start"`
	End            int64  `json:"end"`
	TagKeys        []any  `json:"tag_keys,omitempty"`
	TagKeyError    string `json:"tag_key_error,omitempty"`
	Result         any    `json:"result"`
}

func Metrics(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `Metrics 查询

Usage:
  byte-cli metrics <command>

Commands:
  tagk <metric-name>    查询 metric 的 tag keys
  query <metric-name>   查询 metric 数据
  field <metric-name>   查询 metric 的 fields`)
		return 0, nil
	}
	switch args[0] {
	case "tagk":
		return metricsTagK(args[1:], out)
	case "query":
		return metricsQuery(args[1:], out)
	case "field":
		return metricsField(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown metrics command: %s", args[0])
	}
}

func metricsTagK(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 metric 的 tag keys

Usage:
  byte-cli metrics tagk [options] <metric-name>

Options:
  -r, --region <region>           认证/平台区域 (default: "us")
  --tagkv <key=value>             已选 tag 过滤条件，可多次使用
  --metrics-region <region>       覆盖请求参数 _region
  --tenant <tenant>               租户 (default: "default")
  --endpoint <url>                覆盖 Metrics API base URL
  --origin <url>                  覆盖浏览器 Origin/Referer
  --json                          JSON 格式输出
  -h, --help                      显示帮助信息`)
		return 0, nil
	}
	fs, opts, tagKVValues := newMetricsFlagSet("metrics tagk", out, true, true)
	if err := fs.Parse(normalizeFlags(args, metricsValueFlags(), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("tagk requires METRIC_NAME")
	}
	tagKVs, err := parseMetricsTagKVs(tagKVValues)
	if err != nil {
		return 1, err
	}
	opts.TagKVs = tagKVs
	metricName := fs.Arg(0)
	resp, cfg, err := getMetricsTagK(metricName, *opts)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	printMetricsSuggest(out, "TagK", metricName, "", opts.Region, cfg.PlatformRegion, resp.Data)
	return 0, nil
}

func metricsQuery(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 metric 数据

Usage:
  byte-cli metrics query [options] <metric-name>

Options:
  -r, --region <region>           认证/平台区域 (default: "us")
  --metrics-region <region>       覆盖请求参数 _region
  --tenant <tenant>               租户 (default: "default")
  --endpoint <url>                覆盖 Metrics API base URL
  --origin <url>                  覆盖浏览器 Origin/Referer
  --start <ms>                    查询开始时间，毫秒时间戳
  --end <ms>                      查询结束时间，毫秒时间戳 (default: now)
  --window <duration>             未指定 --start 时的查询窗口 (default: "30m")
  --aggregator <name>             聚合方式 (default: "sum")
  --field <name>                  查询 field/multiFieldExpr (default: "delta")
  --group-by <tagk>               按 tag key 分组，可多次使用
  --filter <key=value>            tag 过滤条件，可多次使用
  --topk <value>                  topK 参数 (default: "top-10-max")
  --downsample <value>            downsample 参数
  --json                          JSON 格式输出
  -h, --help                      显示帮助信息`)
		return 0, nil
	}
	fs, opts, groupByValues, filterValues := newMetricsQueryFlagSet(out)
	if err := fs.Parse(normalizeFlags(args, metricsQueryValueFlags(), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("query requires METRIC_NAME")
	}
	filters, err := parseMetricsTagKVs(filterValues)
	if err != nil {
		return 1, err
	}
	opts.Filters = filters
	opts.GroupBys = groupByValues
	metricName := fs.Arg(0)
	result, err := getMetricsQuery(metricName, *opts, time.Now())
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, result)
	}
	printMetricsQuery(out, result)
	return 0, nil
}

func metricsField(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 metric 的 fields

Usage:
  byte-cli metrics field [options] <metric-name>

Options:
  -r, --region <region>           认证/平台区域 (default: "us")
  --tenant <tenant>               租户 (default: "default")
  --endpoint <url>                覆盖 Metrics API base URL
  --origin <url>                  覆盖浏览器 Origin/Referer
  --json                          JSON 格式输出
  -h, --help                      显示帮助信息`)
		return 0, nil
	}
	fs, opts, _ := newMetricsFlagSet("metrics field", out, false, false)
	if err := fs.Parse(normalizeFlags(args, metricsFieldValueFlags(), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("field requires METRIC_NAME")
	}
	metricName := fs.Arg(0)
	resp, cfg, err := getMetricsField(metricName, *opts)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	printMetricsSuggest(out, "Field", metricName, "", opts.Region, cfg.PlatformRegion, resp.Data)
	return 0, nil
}

func newMetricsFlagSet(name string, out Output, withTagKV, withMetricsRegion bool) (*flag.FlagSet, *metricsOptions, arrayFlags) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out.Err)
	opts := &metricsOptions{Region: "us", Tenant: defaultMetricsTenant}
	tagKVValues := arrayFlags{}
	fs.StringVar(&opts.Region, "region", opts.Region, "认证/平台区域")
	fs.StringVar(&opts.Region, "r", opts.Region, "认证/平台区域")
	if withTagKV {
		fs.Var(&tagKVValues, "tagkv", "已选 tag 过滤条件 key=value")
	}
	if withMetricsRegion {
		fs.StringVar(&opts.MetricsRegion, "metrics-region", "", "覆盖请求参数 _region")
	}
	fs.StringVar(&opts.Tenant, "tenant", opts.Tenant, "租户")
	fs.StringVar(&opts.APIBase, "endpoint", "", "Metrics API base URL")
	fs.StringVar(&opts.Origin, "origin", "", "浏览器 Origin/Referer")
	return fs, opts, tagKVValues
}

func newMetricsQueryFlagSet(out Output) (*flag.FlagSet, *metricsQueryOptions, arrayFlags, arrayFlags) {
	fs := flag.NewFlagSet("metrics query", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	opts := &metricsQueryOptions{
		Metrics:    metricsOptions{Region: "us", Tenant: defaultMetricsTenant},
		Aggregator: "sum",
		Field:      "delta",
		TopK:       "top-10-max",
		Window:     defaultMetricsQueryWindow,
	}
	groupByValues := arrayFlags{}
	filterValues := arrayFlags{}
	fs.StringVar(&opts.Metrics.Region, "region", opts.Metrics.Region, "认证/平台区域")
	fs.StringVar(&opts.Metrics.Region, "r", opts.Metrics.Region, "认证/平台区域")
	fs.StringVar(&opts.Metrics.MetricsRegion, "metrics-region", "", "覆盖请求参数 _region")
	fs.StringVar(&opts.Metrics.Tenant, "tenant", opts.Metrics.Tenant, "租户")
	fs.StringVar(&opts.Metrics.APIBase, "endpoint", "", "Metrics API base URL")
	fs.StringVar(&opts.Metrics.Origin, "origin", "", "浏览器 Origin/Referer")
	fs.StringVar(&opts.Start, "start", "", "查询开始时间，毫秒时间戳")
	fs.StringVar(&opts.End, "end", "", "查询结束时间，毫秒时间戳")
	fs.StringVar(&opts.Window, "window", opts.Window, "未指定 --start 时的查询窗口")
	fs.StringVar(&opts.Aggregator, "aggregator", opts.Aggregator, "聚合方式")
	fs.StringVar(&opts.Field, "field", opts.Field, "查询 field/multiFieldExpr")
	fs.Var(&groupByValues, "group-by", "按 tag key 分组")
	fs.Var(&filterValues, "filter", "tag 过滤条件 key=value")
	fs.StringVar(&opts.TopK, "topk", opts.TopK, "topK 参数")
	fs.StringVar(&opts.Downsample, "downsample", "", "downsample 参数")
	return fs, opts, groupByValues, filterValues
}

func metricsValueFlags() map[string]bool {
	return stringSet("r", "region", "tagkv", "metrics-region", "tenant", "endpoint", "origin")
}

func metricsFieldValueFlags() map[string]bool {
	return stringSet("r", "region", "tenant", "endpoint", "origin")
}

func metricsQueryValueFlags() map[string]bool {
	return stringSet("r", "region", "metrics-region", "tenant", "endpoint", "origin", "start", "end", "window", "aggregator", "field", "group-by", "filter", "topk", "downsample")
}

func getMetricsTagK(metricName string, opts metricsOptions) (MetricsSuggestResponse, metricsRegionConfig, error) {
	cfg, err := resolveMetricsRegion(opts, true)
	if err != nil {
		return MetricsSuggestResponse{}, metricsRegionConfig{}, err
	}
	params := metricsTagKParams(metricName, opts.Tenant, cfg.PlatformRegion)
	for _, tagKV := range opts.TagKVs {
		params.Add("tagkv", tagKV.Key+","+tagKV.Value)
	}
	resp, err := requestMetricsSuggest(cfg, "suggest/tagk", params)
	if err != nil {
		return MetricsSuggestResponse{}, cfg, err
	}
	return resp, cfg, nil
}

func getMetricsField(metricName string, opts metricsOptions) (MetricsSuggestResponse, metricsRegionConfig, error) {
	cfg, err := resolveMetricsRegion(opts, false)
	if err != nil {
		return MetricsSuggestResponse{}, metricsRegionConfig{}, err
	}
	resp, err := requestMetricsSuggest(cfg, "suggest/field", metricsFieldParams(metricName, opts.Tenant))
	if err != nil {
		return MetricsSuggestResponse{}, cfg, err
	}
	return resp, cfg, nil
}

func getMetricsQuery(metricName string, opts metricsQueryOptions, now time.Time) (MetricsQueryOutput, error) {
	cfg, err := resolveMetricsRegion(opts.Metrics, true)
	if err != nil {
		return MetricsQueryOutput{}, err
	}
	start, end, err := resolveMetricsQueryRange(opts.Start, opts.End, opts.Window, now)
	if err != nil {
		return MetricsQueryOutput{}, err
	}
	result, err := requestMetricsRawPost(cfg, metricsQueryAPIBase(cfg), "metrics/query", metricsQueryParams(cfg.PlatformRegion), metricsQueryBody(metricName, opts, start, end))
	if err != nil {
		return MetricsQueryOutput{}, err
	}
	output := MetricsQueryOutput{
		Metric:         metricName,
		Region:         opts.Metrics.Region,
		PlatformRegion: cfg.PlatformRegion,
		Start:          start,
		End:            end,
		Result:         result,
	}
	if tagResp, _, err := getMetricsTagK(metricName, opts.Metrics); err == nil {
		output.TagKeys = metricsSuggestItems(tagResp.Data)
	} else {
		output.TagKeyError = err.Error()
	}
	return output, nil
}

func resolveMetricsRegion(opts metricsOptions, requirePlatformRegion bool) (metricsRegionConfig, error) {
	region := strings.ToLower(strings.TrimSpace(opts.Region))
	if region == "" {
		region = "us"
	}
	cfg, ok := metricsRegionMap[region]
	if !ok {
		if opts.APIBase == "" || (requirePlatformRegion && opts.MetricsRegion == "") {
			hint := "--endpoint"
			if requirePlatformRegion {
				hint = "--endpoint 和 --metrics-region"
			}
			return metricsRegionConfig{}, fmt.Errorf("metrics 未配置区域: %s，可选: us；其他区域可用 %s 指定", opts.Region, hint)
		}
		cfg = metricsRegionConfig{AuthRegion: region}
	}
	if opts.APIBase != "" {
		cfg.APIBase = opts.APIBase
		cfg.QueryAPIBase = opts.APIBase
	}
	if opts.MetricsRegion != "" {
		cfg.PlatformRegion = opts.MetricsRegion
	}
	if opts.Origin != "" {
		cfg.Origin = strings.TrimRight(opts.Origin, "/")
		cfg.Referer = cfg.Origin + "/"
	}
	if cfg.APIBase == "" || (requirePlatformRegion && cfg.PlatformRegion == "") {
		return metricsRegionConfig{}, fmt.Errorf("metrics 区域配置不完整: %s", opts.Region)
	}
	if cfg.AuthRegion == "" {
		cfg.AuthRegion = region
	}
	return cfg, nil
}

func metricsTagKParams(metricName, tenant, platformRegion string) url.Values {
	if tenant == "" {
		tenant = defaultMetricsTenant
	}
	return url.Values{
		"_region":     {platformRegion},
		"_tenant":     {tenant},
		"metric_name": {metricName},
	}
}

func metricsFieldParams(metricName, tenant string) url.Values {
	if tenant == "" {
		tenant = defaultMetricsTenant
	}
	return url.Values{
		"_tenant": {tenant},
		"m":       {metricName},
	}
}

func parseMetricsTagKVs(values []string) ([]metricsTagKV, error) {
	out := make([]metricsTagKV, 0, len(values))
	for _, value := range values {
		key, tagValue, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("tagkv 格式错误: %s，应为 key=value", value)
		}
		out = append(out, metricsTagKV{Key: strings.TrimSpace(key), Value: tagValue})
	}
	return out, nil
}

func metricsQueryAPIBase(cfg metricsRegionConfig) string {
	if cfg.QueryAPIBase != "" {
		return cfg.QueryAPIBase
	}
	return cfg.APIBase
}

func metricsQueryParams(platformRegion string) url.Values {
	return url.Values{
		"_axis":            {"0"},
		"_isGroupByRegion": {"false"},
		"_region":          {platformRegion},
		"_y0max":           {""},
		"_y0min":           {""},
		"_y1max":           {""},
		"_y1min":           {""},
		"json":             {"true"},
		"show_stats":       {"true"},
	}
}

func metricsQueryBody(metricName string, opts metricsQueryOptions, start, end int64) map[string]any {
	filters := make([]map[string]any, 0, len(opts.GroupBys)+len(opts.Filters))
	for _, groupBy := range opts.GroupBys {
		groupBy = strings.TrimSpace(groupBy)
		if groupBy == "" {
			continue
		}
		filters = append(filters, map[string]any{"tagk": groupBy, "filter": "*", "type": "literal_or", "groupBy": true})
	}
	for _, filter := range opts.Filters {
		filters = append(filters, map[string]any{"tagk": filter.Key, "filter": filter.Value, "type": "literal_or", "groupBy": false})
	}
	return map[string]any{
		"allowCoprocessor": false,
		"end":              end,
		"start":            start,
		"queries": []map[string]any{
			{
				"metric":         metricName,
				"tenant":         opts.Metrics.Tenant,
				"aggregator":     opts.Aggregator,
				"rightAxis":      0,
				"rate":           false,
				"rateOptions":    map[string]any{"counter": true, "diff": false, "order": "before_downsample"},
				"topK":           opts.TopK,
				"multiFieldExpr": opts.Field,
				"isMultiField":   true,
				"downsample":     opts.Downsample,
				"filters":        filters,
			},
		},
	}
}

func resolveMetricsQueryRange(startValue, endValue, windowValue string, now time.Time) (int64, int64, error) {
	end := now.UnixMilli()
	if strings.TrimSpace(endValue) != "" {
		value, err := strconv.ParseInt(strings.TrimSpace(endValue), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("--end 格式错误，应为毫秒时间戳: %w", err)
		}
		end = value
	}
	if strings.TrimSpace(startValue) != "" {
		value, err := strconv.ParseInt(strings.TrimSpace(startValue), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("--start 格式错误，应为毫秒时间戳: %w", err)
		}
		return value, end, nil
	}
	if strings.TrimSpace(windowValue) == "" {
		windowValue = defaultMetricsQueryWindow
	}
	window, err := time.ParseDuration(windowValue)
	if err != nil {
		return 0, 0, fmt.Errorf("--window 格式错误: %w", err)
	}
	return end - window.Milliseconds(), end, nil
}

func requestMetricsSuggest(cfg metricsRegionConfig, path string, params url.Values) (MetricsSuggestResponse, error) {
	token, err := tokenFor(cfg.AuthRegion)
	if err != nil {
		return MetricsSuggestResponse{}, err
	}
	resp, err := httpclient.Request(http.MethodGet, metricsSuggestURL(cfg.APIBase, path, params), nil, metricsHeaders(token, cfg, false))
	if err != nil {
		return MetricsSuggestResponse{}, err
	}
	payload, err := readMetricsSuggestResponse(resp)
	if err != nil {
		return MetricsSuggestResponse{}, err
	}
	if payload.Code != 0 {
		return MetricsSuggestResponse{}, fmt.Errorf("metrics API error code %d: %s", payload.Code, payload.Message)
	}
	return payload, nil
}

func requestMetricsRawPost(cfg metricsRegionConfig, apiBase, path string, params url.Values, body any) (any, error) {
	token, err := tokenFor(cfg.AuthRegion)
	if err != nil {
		return nil, err
	}
	resp, err := httpclient.Request(http.MethodPost, metricsSuggestURL(apiBase, path, params), body, metricsHeaders(token, cfg, true))
	if err != nil {
		return nil, err
	}
	return readMetricsRawResponse(resp)
}

func metricsSuggestURL(apiBase, path string, params url.Values) string {
	return strings.TrimRight(apiBase, "/") + "/" + strings.TrimLeft(path, "/") + "?" + params.Encode()
}

func metricsHeaders(token string, cfg metricsRegionConfig, withJSONBody bool) map[string]string {
	fetchSite := cfg.FetchSite
	if fetchSite == "" {
		fetchSite = "cross-site"
	}
	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Authorization":      token,
		"Connection":         "keep-alive",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     fetchSite,
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36",
		"accept-language":    "zh",
		"sec-ch-ua":          `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"macOS"`,
	}
	if withJSONBody {
		headers["Content-Type"] = "application/json;charset=UTF-8"
	}
	if cfg.Origin != "" {
		origin := strings.TrimRight(cfg.Origin, "/")
		headers["Origin"] = origin
		if cfg.Referer != "" {
			headers["Referer"] = cfg.Referer
		} else {
			headers["Referer"] = origin + "/"
		}
	}
	return headers
}

func readMetricsSuggestResponse(resp *http.Response) (MetricsSuggestResponse, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return MetricsSuggestResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return MetricsSuggestResponse{}, fmt.Errorf("HTTP %d%s", resp.StatusCode, metricsErrorDetail(resp, data))
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return MetricsSuggestResponse{}, err
	}
	if object, ok := raw.(map[string]any); ok {
		if _, hasData := object["data"]; hasData {
			var payload MetricsSuggestResponse
			if err := json.Unmarshal(data, &payload); err != nil {
				return MetricsSuggestResponse{}, err
			}
			return payload, nil
		}
		if _, hasCode := object["code"]; hasCode {
			var payload MetricsSuggestResponse
			if err := json.Unmarshal(data, &payload); err != nil {
				return MetricsSuggestResponse{}, err
			}
			return payload, nil
		}
	}
	return MetricsSuggestResponse{Code: 0, Data: raw}, nil
}

func readMetricsRawResponse(resp *http.Response) (any, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d%s", resp.StatusCode, metricsErrorDetail(resp, data))
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func metricsErrorDetail(resp *http.Response, data []byte) string {
	parts := []string{}
	if logID := resp.Header.Get("x-tt-logid"); logID != "" {
		parts = append(parts, "logid="+logID)
	}
	if body := strings.TrimSpace(string(data)); body != "" {
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		parts = append(parts, body)
	}
	if len(parts) == 0 {
		return ""
	}
	return ": " + strings.Join(parts, "; ")
}

func printMetricsSuggest(out Output, label, metricName, tagKey, region, platformRegion string, data any) {
	items := metricsSuggestItems(data)
	fmt.Fprintf(out.Out, "Metric: %s\n", metricName)
	if tagKey != "" {
		fmt.Fprintf(out.Out, "TagK: %s\n", tagKey)
	}
	if platformRegion != "" {
		fmt.Fprintf(out.Out, "区域: %s (%s)\n", region, platformRegion)
	} else {
		fmt.Fprintf(out.Out, "区域: %s\n", region)
	}
	fmt.Fprintf(out.Out, "%s 数量: %d\n\n", label, len(items))
	for _, item := range items {
		fmt.Fprintf(out.Out, "  %s\n", formatMetricsValue(item))
	}
}

func metricsSuggestItems(data any) []any {
	if data == nil {
		return nil
	}
	if items, ok := data.([]any); ok {
		return items
	}
	return []any{data}
}

func printMetricsQuery(out Output, result MetricsQueryOutput) {
	fmt.Fprintf(out.Out, "Metric: %s\n", result.Metric)
	fmt.Fprintf(out.Out, "区域: %s (%s)\n", result.Region, result.PlatformRegion)
	fmt.Fprintf(out.Out, "时间: %d - %d\n", result.Start, result.End)
	if result.TagKeyError != "" {
		fmt.Fprintf(out.Out, "TagK 查询失败: %s\n", result.TagKeyError)
	} else {
		fmt.Fprintf(out.Out, "TagK 数量: %d\n", len(result.TagKeys))
		for _, key := range result.TagKeys {
			fmt.Fprintf(out.Out, "  %s\n", formatMetricsValue(key))
		}
	}
	fmt.Fprintln(out.Out, "\nResult:")
	printMetricsPlainValue(out, result.Result, 0)
}

func printMetricsPlainValue(out Output, value any, indent int) {
	items := metricsSuggestItems(value)
	if len(items) > 1 {
		for _, item := range items {
			fmt.Fprintf(out.Out, "%s%s\n", strings.Repeat(" ", indent), formatMetricsValue(item))
		}
		return
	}
	fmt.Fprintf(out.Out, "%s%s\n", strings.Repeat(" ", indent), formatMetricsValue(value))
}

func formatMetricsValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}
