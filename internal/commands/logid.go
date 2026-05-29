package commands

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/DreamCats/byte-cli/internal/httpclient"
)

var logServiceURLs = map[string]string{
	"us":   "https://logservice-tx.tiktok-us.org/streamlog/platform/microservice/v1/query/trace",
	"i18n": "https://logservice-sg.tiktok-row.org/streamlog/platform/microservice/v1/query/trace",
	"eu":   "https://logservice-eu-ttp.tiktok-eu.org/streamlog/platform/microservice/v1/query/trace",
}

var vregionMap = map[string]string{
	"us":   "US-TTP,US-TTP2",
	"i18n": "Singapore-Central",
	"eu":   "EU-TTP2,US-EastRed",
}

type LogQueryResponse struct {
	Data LogData `json:"data"`
}

type LogData struct {
	Items []LogItem `json:"items"`
}

type LogItem struct {
	ID    string     `json:"id"`
	Group LogGroup   `json:"group"`
	Value []LogValue `json:"value"`
}

type LogGroup struct {
	PSM     string `json:"psm"`
	PodName string `json:"pod_name"`
	IPv4    string `json:"ipv4"`
	Env     string `json:"env"`
	VRegion string `json:"vregion"`
	IDC     string `json:"idc"`
}

type LogValue struct {
	KVList []KVPair `json:"kv_list"`
	Level  string   `json:"level"`
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FlattenedLogEntry struct {
	PSM      string `json:"psm"`
	PodName  string `json:"pod_name"`
	Level    string `json:"level"`
	Msg      string `json:"msg"`
	Location string `json:"location"`
	Env      string `json:"env"`
	VRegion  string `json:"vregion"`
}

type logQueryOptions struct {
	PSMList     []string
	Keywords    []string
	Levels      []string
	MaxLen      int
	ScanSpanMin int
}

func LogID(args []string, out Output) (int, error) {
	if len(args) == 0 || wantsHelp(args) {
		printHelp(out, `通过 Log ID 查询分布式日志链路

Usage:
  byte-cli logid [options] <trace-id>

Options:
  -r, --region <region>    区域: cn/i18n/us/eu (default: "cn")
  -p, --psm <psm...>       PSM 服务名过滤（可多次指定）
  -k, --keyword <keyword...> 关键词过滤（可多次指定，OR 关系）
  -l, --level <level...>   日志级别过滤（可多次指定，如 error warn）
  --limit <n>              最多显示条数，默认 20 (default: "20")
  --max-len <len>          消息最大长度，默认 300，0 表示不截断 (default: "300")
  --json                   JSON 格式输出（不限制条数和长度）
  -h, --help               显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("logid", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	region := fs.String("region", "cn", "区域: cn/i18n/us/eu")
	fs.StringVar(region, "r", "cn", "区域: cn/i18n/us/eu")
	psms := arrayFlags{}
	keywords := arrayFlags{}
	levels := arrayFlags{}
	fs.Var(&psms, "psm", "PSM 服务名过滤")
	fs.Var(&psms, "p", "PSM 服务名过滤")
	fs.Var(&keywords, "keyword", "关键词过滤")
	fs.Var(&keywords, "k", "关键词过滤")
	fs.Var(&levels, "level", "日志级别过滤")
	fs.Var(&levels, "l", "日志级别过滤")
	limitOpt := fs.String("limit", "20", "最多显示条数")
	maxLenOpt := fs.String("max-len", "300", "消息最大长度")
	if err := fs.Parse(normalizeFlags(args, stringSet("r", "region", "p", "psm", "k", "keyword", "l", "level", "limit", "max-len"), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("logid requires TRACE_ID")
	}
	token, err := tokenFor(*region)
	if err != nil {
		return authExitCode(err), err
	}
	maxLen, _ := strconv.Atoi(*maxLenOpt)
	if out.JSON {
		maxLen = 0
	}
	entries, err := queryLogs(fs.Arg(0), *region, token, logQueryOptions{
		PSMList:  psms,
		Keywords: keywords,
		Levels:   levels,
		MaxLen:   maxLen,
	})
	if err != nil {
		return 1, err
	}
	total := len(entries)
	limit, _ := strconv.Atoi(*limitOpt)
	if out.JSON {
		return printPrettyJSON(out, entries)
	}
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	printLogEntries(entries, total, out)
	return 0, nil
}

func queryLogs(logid, region, token string, options logQueryOptions) ([]FlattenedLogEntry, error) {
	url := logServiceURLs[region]
	if url == "" {
		return nil, fmt.Errorf("区域 %s 暂不支持日志查询（仅 us/i18n/eu）", region)
	}
	scanSpan := options.ScanSpanMin
	if scanSpan == 0 {
		scanSpan = 10
	}
	body := map[string]any{
		"logid":            logid,
		"psm_list":         options.PSMList,
		"scan_span_in_min": scanSpan,
		"vregion":          vregionMap[region],
	}
	resp, err := httpclient.Post(url, body, map[string]string{"X-Jwt-Token": token})
	if err != nil {
		return nil, err
	}
	var payload LogQueryResponse
	if err := readJSON(resp, &payload); err != nil {
		return nil, err
	}
	sanitizer := newMessageSanitizer()
	levels := make([]string, 0, len(options.Levels))
	for _, level := range options.Levels {
		levels = append(levels, strings.ToLower(level))
	}
	results := []FlattenedLogEntry{}
	for _, item := range payload.Data.Items {
		results = append(results, extractEntries(item, sanitizer, options.Keywords, levels, options.MaxLen)...)
	}
	return results, nil
}

func extractEntries(item LogItem, sanitizer messageSanitizer, keywords, levels []string, maxLen int) []FlattenedLogEntry {
	entries := []FlattenedLogEntry{}
	for _, value := range item.Value {
		msg := ""
		location := ""
		for _, kv := range value.KVList {
			switch kv.Key {
			case "_msg":
				msg = kv.Value
			case "_location":
				location = kv.Value
			}
		}
		if msg == "" {
			continue
		}
		if len(levels) > 0 && !stringIn(strings.ToLower(value.Level), levels) {
			continue
		}
		msg = sanitizer.sanitize(msg)
		if !keywordMatches(msg, keywords) {
			continue
		}
		if maxLen > 0 && len(msg) > maxLen {
			msg = msg[:maxLen] + "..."
		}
		entries = append(entries, FlattenedLogEntry{
			PSM: item.Group.PSM, PodName: item.Group.PodName, Level: value.Level,
			Msg: msg, Location: location, Env: item.Group.Env, VRegion: item.Group.VRegion,
		})
	}
	return entries
}

func printLogEntries(entries []FlattenedLogEntry, total int, out Output) {
	if len(entries) == 0 {
		fmt.Fprintln(out.Out, "未找到日志条目")
		return
	}
	if total > len(entries) {
		fmt.Fprintf(out.Out, "共 %d 条日志（显示前 %d 条，--limit 0 显示全部）:\n\n", total, len(entries))
	} else {
		fmt.Fprintf(out.Out, "共 %d 条日志:\n\n", len(entries))
	}
	for i, entry := range entries {
		fmt.Fprintf(out.Out, "[%d] %s  %s  %s\n", i+1, entry.Level, entry.PSM, entry.VRegion)
		if entry.Location != "" {
			fmt.Fprintf(out.Out, "    位置: %s\n", entry.Location)
		}
		fmt.Fprintf(out.Out, "    %s\n\n", entry.Msg)
	}
}

type messageSanitizer struct {
	patterns []*regexp.Regexp
}

func newMessageSanitizer() messageSanitizer {
	raw := []string{
		`_compliance_nlp_log`,
		`_compliance_whitelist_log`,
		`_compliance_source=footprint`,
		`"user_extra":\s*"[\s\S]*?"`,
		`"LogID":\s*"[^"]*"`,
		`"Addr":\s*"[^"]*"`,
		`"Client":\s*"[^"]*"`,
		`\{\{rip=[^}]*\}\}`,
	}
	patterns := make([]*regexp.Regexp, 0, len(raw))
	for _, expr := range raw {
		patterns = append(patterns, regexp.MustCompile(expr))
	}
	return messageSanitizer{patterns: patterns}
}

func (s messageSanitizer) sanitize(text string) string {
	result := text
	for _, pattern := range s.patterns {
		result = pattern.ReplaceAllString(result, "")
	}
	return cleanWhitespace(result)
}

func keywordMatches(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func cleanWhitespace(text string) string {
	space := regexp.MustCompile(`[ \t]{2,}`)
	newlines := regexp.MustCompile(`\n{3,}`)
	text = space.ReplaceAllString(text, " ")
	text = newlines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
