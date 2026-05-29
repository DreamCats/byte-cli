package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DreamCats/byte-cli/internal/httpclient"
)

var mcpDomainMap = map[string]string{
	"cn":   "cloud.bytedance.net",
	"us":   "cloud.tiktok-us.net",
	"eu":   "cloud-eu.tiktok-row.net",
	"i18n": "cloud.tiktok-row.net",
}

var mcpServerSuffixMap = map[string]string{
	"cn":   "mcp.bytedance.net",
	"us":   "mcp-usttp.tiktok-us.net",
	"eu":   "mcp-eu.tiktok-row.net",
	"i18n": "mcp.tiktok-row.net",
}

const mcpServersPath = "/api/v1/aipaas/api/v1/mcp/servers"

type MCPServer struct {
	ServerID          string   `json:"server_id"`
	PSM               string   `json:"psm"`
	EnvName           string   `json:"env_name"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Owner             string   `json:"owner"`
	Subscribers       []string `json:"subscribers"`
	CurrentRevisionID string   `json:"current_revision_id"`
	AuthEnabled       bool     `json:"auth_enabled"`
	AllowedPSMs       []string `json:"allowed_psms"`
	Admins            []string `json:"admins"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

type MCPServerListResponse struct {
	Code  int         `json:"code"`
	Error string      `json:"error"`
	Data  []MCPServer `json:"data"`
}

type ToolParameter struct {
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	Format      string                   `json:"format,omitempty"`
	Properties  map[string]ToolParameter `json:"properties"`
	Items       *ToolParameter           `json:"items"`
	Required    []string                 `json:"required"`
}

type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema ToolParameter     `json:"inputSchema"`
	Annotations map[string]string `json:"annotations"`
}

type ToolListResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Tools []Tool `json:"tools"`
	} `json:"result"`
}

type ToolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCallResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Content []ToolCallContent `json:"content"`
		Meta    map[string]any    `json:"meta"`
	} `json:"result"`
}

func MCP(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `MCP Server 查询

Usage:
  byte-cli mcp <command>

Commands:
  list                         查询 MCP Server 列表
  call <server-id> <tool-name> 调用 MCP Server 的工具
  tools <server-id>            查询 MCP Server 的工具列表`)
		return 0, nil
	}
	switch args[0] {
	case "list":
		return mcpList(args[1:], out)
	case "tools":
		return mcpTools(args[1:], out)
	case "call":
		return mcpCall(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown mcp command: %s", args[0])
	}
}

func mcpList(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 MCP Server 列表

Usage:
  byte-cli mcp list [options]

Options:
  -s, --search <search>  按 PSM 搜索
  -e, --env <env>        环境名称 (default: "prod")
  -r, --region <region>  区域 (default: "cn")
  -l, --limit <limit>    每页数量 (default: "10")
  -o, --offset <offset>  偏移量 (default: "0")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("mcp list", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	search := fs.String("search", "", "按 PSM 搜索")
	fs.StringVar(search, "s", "", "按 PSM 搜索")
	env := fs.String("env", "prod", "环境名称")
	fs.StringVar(env, "e", "prod", "环境名称")
	region := fs.String("region", "cn", "区域")
	fs.StringVar(region, "r", "cn", "区域")
	limit := fs.String("limit", "10", "每页数量")
	fs.StringVar(limit, "l", "10", "每页数量")
	offset := fs.String("offset", "0", "偏移量")
	fs.StringVar(offset, "o", "0", "偏移量")
	if err := fs.Parse(normalizeFlags(args, stringSet("s", "search", "e", "env", "r", "region", "l", "limit", "o", "offset"), nil)); err != nil {
		return 1, err
	}
	limitNum, _ := strconv.Atoi(*limit)
	offsetNum, _ := strconv.Atoi(*offset)
	resp, err := getMcpServers(*search, *env, *region, limitNum, offsetNum)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	if len(resp.Data) == 0 {
		fmt.Fprintln(out.Out, "未找到 MCP Server")
		return 0, nil
	}
	fmt.Fprintf(out.Out, "区域: %s\n", *region)
	fmt.Fprintf(out.Out, "环境: %s\n", *env)
	fmt.Fprintf(out.Out, "数量: %d\n\n", len(resp.Data))
	for _, server := range resp.Data {
		fmt.Fprintf(out.Out, "  %s (ID: %s)\n", server.Name, server.ServerID)
		fmt.Fprintf(out.Out, "    PSM:        %s\n", server.PSM)
		if server.Description != "" {
			fmt.Fprintf(out.Out, "    描述:       %s\n", server.Description)
		}
		fmt.Fprintf(out.Out, "    负责人:     %s\n", server.Owner)
		fmt.Fprintf(out.Out, "    管理员:     %s\n", strings.Join(server.Admins, ", "))
		fmt.Fprintf(out.Out, "    认证:       %s\n", enabledDisabled(server.AuthEnabled))
		fmt.Fprintf(out.Out, "    版本:       %s\n", server.CurrentRevisionID)
		fmt.Fprintf(out.Out, "    更新时间:   %s\n\n", formatServerDate(server.UpdatedAt))
	}
	return 0, nil
}

func mcpTools(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 MCP Server 的工具列表

Usage:
  byte-cli mcp tools [options] <server-id>

Options:
  -r, --region <region>  区域 (default: "cn")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("mcp tools", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	region := fs.String("region", "cn", "区域")
	fs.StringVar(region, "r", "cn", "区域")
	if err := fs.Parse(normalizeFlags(args, stringSet("r", "region"), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("tools requires SERVER_ID")
	}
	resp, err := getMcpTools(fs.Arg(0), *region)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	tools := resp.Result.Tools
	if len(tools) == 0 {
		fmt.Fprintln(out.Out, "未找到工具")
		return 0, nil
	}
	fmt.Fprintf(out.Out, "Server ID: %s\n", fs.Arg(0))
	fmt.Fprintf(out.Out, "区域: %s\n", *region)
	fmt.Fprintf(out.Out, "工具数量: %d\n\n", len(tools))
	for _, tool := range tools {
		fmt.Fprintf(out.Out, "  %s\n", tool.Name)
		if tool.Description != "" {
			fmt.Fprintf(out.Out, "    描述: %s\n", tool.Description)
		}
		if len(tool.InputSchema.Properties) > 0 {
			fmt.Fprintln(out.Out, "    参数:")
			for name, param := range tool.InputSchema.Properties {
				required := ""
				if stringIn(name, tool.InputSchema.Required) {
					required = " [必填]"
				}
				desc := param.Description
				if desc == "" {
					desc = param.Type
				}
				fmt.Fprintf(out.Out, "      - %s: %s%s\n", name, desc, required)
			}
		}
		fmt.Fprintln(out.Out)
	}
	return 0, nil
}

func mcpCall(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `调用 MCP Server 的工具

Usage:
  byte-cli mcp call [options] <server-id> <tool-name>

Options:
  -a, --arg <arg...>     工具参数，格式: key=value（可多次使用）
  -r, --region <region>  区域 (default: "cn")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("mcp call", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	argValues := arrayFlags{}
	fs.Var(&argValues, "arg", "工具参数 key=value")
	fs.Var(&argValues, "a", "工具参数 key=value")
	region := fs.String("region", "cn", "区域")
	fs.StringVar(region, "r", "cn", "区域")
	if err := fs.Parse(normalizeFlags(args, stringSet("a", "arg", "r", "region"), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() < 2 {
		return 1, fmt.Errorf("call requires SERVER_ID TOOL_NAME")
	}
	arguments := map[string]any{}
	for _, item := range argValues {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return 1, fmt.Errorf("参数格式错误: %s，应为 key=value", item)
		}
		var parsed any
		if err := decodeJSONUseNumber(value, &parsed); err != nil {
			parsed = value
		}
		arguments[key] = parsed
	}
	resp, err := callMcpTool(fs.Arg(0), fs.Arg(1), arguments, *region)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	fmt.Fprintf(out.Out, "Server ID: %s\n", fs.Arg(0))
	fmt.Fprintf(out.Out, "工具: %s\n", fs.Arg(1))
	fmt.Fprintf(out.Out, "区域: %s\n\n", *region)
	if len(resp.Result.Content) == 0 {
		fmt.Fprintln(out.Out, "无返回内容")
		return 0, nil
	}
	for _, content := range resp.Result.Content {
		if content.Type != "text" {
			continue
		}
		var parsed any
		if err := decodeJSONUseNumber(content.Text, &parsed); err == nil {
			_ = out.PrintJSON(parsed)
		} else {
			fmt.Fprintln(out.Out, content.Text)
		}
	}
	return 0, nil
}

func getMcpServers(search, env, region string, limit, offset int) (MCPServerListResponse, error) {
	domain := mcpDomainMap[region]
	if domain == "" {
		return MCPServerListResponse{}, fmt.Errorf("MCP 不支持区域: %s，可选: cn/us/eu/i18n", region)
	}
	token, err := tokenFor(region)
	if err != nil {
		return MCPServerListResponse{}, err
	}
	params := url.Values{
		"env":           {env},
		"limit":         {strconv.Itoa(limit)},
		"offset":        {strconv.Itoa(offset)},
		"search":        {search},
		"search_type":   {"own"},
		"sort_by":       {"-updated_at"},
		"search_fields": {"psm"},
	}
	resp, err := httpclient.Get("https://"+domain+mcpServersPath+"?"+params.Encode(), map[string]string{
		"x-jwt-token":           token,
		"x-og-common-path-mode": "true",
		"Accept":                "application/json, text/plain, */*",
		"Accept-Language":       "zh",
	})
	if err != nil {
		return MCPServerListResponse{}, err
	}
	var payload MCPServerListResponse
	return payload, readJSON(resp, &payload)
}

func getMcpTools(serverID, region string) (ToolListResponse, error) {
	var payload ToolListResponse
	err := mcpJSONRPC(serverID, region, map[string]any{"method": "tools/list", "params": map[string]any{}, "jsonrpc": "2.0", "id": 1}, &payload)
	return payload, err
}

func callMcpTool(serverID, toolName string, arguments map[string]any, region string) (ToolCallResponse, error) {
	var payload ToolCallResponse
	body := map[string]any{
		"method":  "tools/call",
		"params":  map[string]any{"name": toolName, "arguments": arguments},
		"jsonrpc": "2.0",
		"id":      1,
	}
	err := mcpJSONRPC(serverID, region, body, &payload)
	return payload, err
}

func mcpJSONRPC(serverID, region string, body map[string]any, dst any) error {
	suffix := mcpServerSuffixMap[region]
	if suffix == "" {
		return fmt.Errorf("MCP 不支持区域: %s，可选: cn/us/eu/i18n", region)
	}
	token, err := tokenFor(region)
	if err != nil {
		return err
	}
	resp, err := httpclient.Post("https://"+serverID+"."+suffix+"/mcp", body, map[string]string{
		"X-Jwt-Token":            token,
		"X-Mcp-Internal-Request": "true",
	})
	if err != nil {
		return err
	}
	return readRawJSON(resp, dst)
}

func decodeJSONUseNumber(text string, dst any) error {
	dec := json.NewDecoder(bytes.NewReader([]byte(text)))
	dec.UseNumber()
	return dec.Decode(dst)
}

type arrayFlags []string

func (a *arrayFlags) String() string { return strings.Join(*a, ",") }
func (a *arrayFlags) Set(v string) error {
	*a = append(*a, v)
	return nil
}

func enabledDisabled(v bool) string {
	if v {
		return "启用"
	}
	return "禁用"
}

func formatServerDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006/01/02 15:04")
}

func stringIn(value string, values []string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
