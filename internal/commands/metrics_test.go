package commands

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestResolveMetricsRegionUsesUSDefaults(t *testing.T) {
	cfg, err := resolveMetricsRegion(metricsOptions{Region: "us"}, true)
	if err != nil {
		t.Fatalf("resolveMetricsRegion returned error: %v", err)
	}
	if cfg.AuthRegion != "us" {
		t.Fatalf("AuthRegion = %q, want us", cfg.AuthRegion)
	}
	if cfg.PlatformRegion != "US-TTP" {
		t.Fatalf("PlatformRegion = %q, want US-TTP", cfg.PlatformRegion)
	}
	if cfg.APIBase != "https://metrics-svc-platform-ttp.tiktok-us.org/byteplot/api/v2" {
		t.Fatalf("unexpected APIBase: %q", cfg.APIBase)
	}
	if cfg.Origin != "https://metrics-fe-ttp-us.tiktok-row.org" {
		t.Fatalf("unexpected Origin: %q", cfg.Origin)
	}
}

func TestResolveMetricsRegionUsesI18NDefaults(t *testing.T) {
	cfg, err := resolveMetricsRegion(metricsOptions{Region: "i18n"}, true)
	if err != nil {
		t.Fatalf("resolveMetricsRegion returned error: %v", err)
	}
	if cfg.AuthRegion != "i18n" {
		t.Fatalf("AuthRegion = %q, want i18n", cfg.AuthRegion)
	}
	if cfg.PlatformRegion != "Singapore-Central" {
		t.Fatalf("PlatformRegion = %q, want Singapore-Central", cfg.PlatformRegion)
	}
	if cfg.APIBase != "https://metrics-fe-i18n.tiktok-row.org/byteplot/api" {
		t.Fatalf("unexpected APIBase: %q", cfg.APIBase)
	}
	if cfg.FetchSite != "same-origin" {
		t.Fatalf("FetchSite = %q, want same-origin", cfg.FetchSite)
	}
}

func TestResolveMetricsRegionAllowsExplicitOverride(t *testing.T) {
	cfg, err := resolveMetricsRegion(metricsOptions{
		Region:        "eu",
		MetricsRegion: "EU-TTP",
		APIBase:       "https://metrics.example.test/api",
		Origin:        "https://metrics-fe.example.test/",
	}, true)
	if err != nil {
		t.Fatalf("resolveMetricsRegion returned error: %v", err)
	}
	if cfg.AuthRegion != "eu" || cfg.PlatformRegion != "EU-TTP" {
		t.Fatalf("unexpected region config: %#v", cfg)
	}
	if cfg.Origin != "https://metrics-fe.example.test" {
		t.Fatalf("Origin = %q, want trimmed origin", cfg.Origin)
	}
}

func TestMetricsSuggestURL(t *testing.T) {
	params := metricsTagKParams("ttec.industry.solution.sales_mask_rule_match", "default", "US-TTP")
	params.Add("tagkv", "result,hit")
	got := metricsSuggestURL("https://metrics.example.test/api/", "/suggest/tagk", params)
	want := "https://metrics.example.test/api/suggest/tagk?_region=US-TTP&_tenant=default&metric_name=ttec.industry.solution.sales_mask_rule_match&tagkv=result%2Chit"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMetricsFieldURL(t *testing.T) {
	got := metricsSuggestURL("https://metrics.example.test/api/", "/suggest/field", metricsFieldParams("ttec.industry.solution.sales_mask_rule_match", "default"))
	want := "https://metrics.example.test/api/suggest/field?_tenant=default&m=ttec.industry.solution.sales_mask_rule_match"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMetricsQueryURL(t *testing.T) {
	got := metricsSuggestURL("https://metrics.example.test/api/", "/metrics/query", metricsQueryParams("US-TTP"))
	want := "https://metrics.example.test/api/metrics/query?_axis=0&_isGroupByRegion=false&_region=US-TTP&_y0max=&_y0min=&_y1max=&_y1min=&json=true&show_stats=true"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMetricsQueryBody(t *testing.T) {
	body := metricsQueryBody("ttec.industry.solution.product_change_event_callback", metricsQueryOptions{
		Metrics:    metricsOptions{Tenant: "default"},
		Aggregator: "sum",
		Field:      "delta",
		TopK:       "top-10-max",
		GroupBys:   []string{"topic"},
		Filters:    []metricsTagKV{{Key: "status", Value: "success"}},
	}, 1000, 2000)
	queries, ok := body["queries"].([]map[string]any)
	if !ok || len(queries) != 1 {
		t.Fatalf("unexpected queries: %#v", body["queries"])
	}
	query := queries[0]
	if query["metric"] != "ttec.industry.solution.product_change_event_callback" {
		t.Fatalf("unexpected metric: %#v", query["metric"])
	}
	filters, ok := query["filters"].([]map[string]any)
	if !ok || len(filters) != 2 {
		t.Fatalf("unexpected filters: %#v", query["filters"])
	}
	if filters[0]["tagk"] != "topic" || filters[0]["groupBy"] != true {
		t.Fatalf("unexpected group by filter: %#v", filters[0])
	}
	if filters[1]["tagk"] != "status" || filters[1]["filter"] != "success" {
		t.Fatalf("unexpected value filter: %#v", filters[1])
	}
}

func TestMetricsHeadersForI18NPost(t *testing.T) {
	cfg := metricsRegionMap["i18n"]
	headers := metricsHeaders("token", cfg, true)
	if headers["Sec-Fetch-Site"] != "same-origin" {
		t.Fatalf("Sec-Fetch-Site = %q, want same-origin", headers["Sec-Fetch-Site"])
	}
	if headers["Content-Type"] != "application/json;charset=UTF-8" {
		t.Fatalf("Content-Type = %q, want json charset", headers["Content-Type"])
	}
	if headers["Referer"] != "https://metrics-fe-i18n.tiktok-row.org/web/plot/metrics" {
		t.Fatalf("Referer = %q", headers["Referer"])
	}
}

func TestMetricsQueryParsesRegionFlag(t *testing.T) {
	fs, opts, _, _ := newMetricsQueryFlagSet(NewOutput(false))
	if err := fs.Parse(normalizeFlags([]string{
		"-r", "invalid",
		"--endpoint", "https://metrics.example.test/api",
		"--metrics-region", "Singapore-Central",
		"metric.name",
	}, metricsQueryValueFlags(), nil)); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if opts.Metrics.Region != "invalid" {
		t.Fatalf("opts.Metrics.Region = %q, want invalid", opts.Metrics.Region)
	}

	out := NewOutput(false)
	code, err := Metrics([]string{
		"query",
		"-r", "invalid",
		"--endpoint", "https://metrics.example.test/api",
		"--metrics-region", "Singapore-Central",
		"metric.name",
	}, out)
	if code == 0 || err == nil {
		t.Fatalf("expected invalid region error, got code=%d err=%v", code, err)
	}
	if err.Error() != "未知区域: invalid，可选: cn, i18n, us, eu, codebase" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveMetricsRegionForFieldAllowsEndpointOnly(t *testing.T) {
	cfg, err := resolveMetricsRegion(metricsOptions{
		Region:  "eu",
		APIBase: "https://metrics.example.test/api",
	}, false)
	if err != nil {
		t.Fatalf("resolveMetricsRegion returned error: %v", err)
	}
	if cfg.AuthRegion != "eu" || cfg.APIBase != "https://metrics.example.test/api" {
		t.Fatalf("unexpected region config: %#v", cfg)
	}
}

func TestParseMetricsTagKVs(t *testing.T) {
	got, err := parseMetricsTagKVs([]string{"result=hit", "scope=global=all"})
	if err != nil {
		t.Fatalf("parseMetricsTagKVs returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d tag kvs, want 2", len(got))
	}
	if got[0].Key != "result" || got[0].Value != "hit" {
		t.Fatalf("unexpected first tag kv: %#v", got[0])
	}
	if got[1].Key != "scope" || got[1].Value != "global=all" {
		t.Fatalf("unexpected second tag kv: %#v", got[1])
	}
}

func TestFormatMetricsValue(t *testing.T) {
	if got := formatMetricsValue("plain"); got != "plain" {
		t.Fatalf("got %q want plain", got)
	}
	got := formatMetricsValue(map[string]any{"k": "v"})
	if got != `{"k":"v"}` {
		t.Fatalf("got %q want JSON object", got)
	}
}

func TestReadMetricsSuggestResponseAllowsRawArray(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`[{"dry_run":["false"]}]`)),
	}
	got, err := readMetricsSuggestResponse(resp)
	if err != nil {
		t.Fatalf("readMetricsSuggestResponse returned error: %v", err)
	}
	items := metricsSuggestItems(got.Data)
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestResolveMetricsQueryRangeUsesWindow(t *testing.T) {
	now := time.UnixMilli(2000)
	start, end, err := resolveMetricsQueryRange("", "", "1s", now)
	if err != nil {
		t.Fatalf("resolveMetricsQueryRange returned error: %v", err)
	}
	if start != 1000 || end != 2000 {
		t.Fatalf("got start=%d end=%d", start, end)
	}
}
