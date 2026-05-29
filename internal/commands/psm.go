package commands

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/DreamCats/byte-cli/internal/auth"
	"github.com/DreamCats/byte-cli/internal/httpclient"
)

const (
	idlInfoURL  = "https://cloud.bytedance.net/api/v1/overpass/api/v3/overpass/platform/idl_info/get_latest_idl_info"
	apiListURL  = "https://cloud.bytedance.net/api/v1/bam/endpoint/list"
	idlRepoBase = "https://code.bytedance.net"
)

var psmDomainMap = map[string]string{
	"us":   "cloud-ttp-us.bytedance.net",
	"eu":   "cloud-eu.tiktok-row.net",
	"i18n": "cloud.tiktok-row.net",
}

type IDLInfo struct {
	PSM           string `json:"psm"`
	RepoName      string `json:"repo_name"`
	IDLPath       string `json:"idl_path"`
	DefaultBranch string `json:"default_branch"`
	IDLVersion    int    `json:"idl_version"`
}

type Endpoint struct {
	Name       string `json:"name"`
	Note       string `json:"note"`
	Owner      string `json:"owner"`
	Version    string `json:"version"`
	Serializer string `json:"serializer"`
	Oneway     bool   `json:"oneway"`
	ModifyTime int64  `json:"modify_time"`
}

func PSM(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `PSM 服务信息查询

Usage:
  byte-cli psm <command>

Commands:
  idl <psm-name>        查询 PSM 的 IDL 信息
  api-list <psm-name>   查询 PSM 的 API 接口列表
  links <psm-name>      生成 PSM 的各平台链接`)
		return 0, nil
	}
	switch args[0] {
	case "idl":
		return psmIDL(args[1:], out)
	case "api-list":
		return psmAPIList(args[1:], out)
	case "links":
		return psmLinks(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown psm command: %s", args[0])
	}
}

func psmIDL(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 PSM 的 IDL 信息

Usage:
  byte-cli psm idl [options] <psm-name>

Options:
  --json       JSON 格式输出
  -h, --help   显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("psm idl", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("idl requires PSM")
	}
	info, err := getIdlInfo(fs.Arg(0))
	if err != nil {
		return 1, err
	}
	if out.JSON {
		return printPrettyJSON(out, info)
	}
	fmt.Fprintf(out.Out, "PSM:            %s\n", info.PSM)
	fmt.Fprintf(out.Out, "RepoName:       %s\n", info.RepoName)
	fmt.Fprintf(out.Out, "IDLPath:        %s\n", info.IDLPath)
	fmt.Fprintf(out.Out, "DefaultBranch:  %s\n", info.DefaultBranch)
	fmt.Fprintf(out.Out, "IDLVersion:     %d\n", info.IDLVersion)
	fmt.Fprintf(out.Out, "IDLRepoURL:     %s\n", idlRepoURL(info))
	fmt.Fprintf(out.Out, "IDLRepoMainIDLURL: %s\n", idlRepoMainURL(info))
	return 0, nil
}

func psmAPIList(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询 PSM 的 API 接口列表

Usage:
  byte-cli psm api-list [options] <psm-name>

Options:
  --json       JSON 格式输出
  -h, --help   显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("psm api-list", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("api-list requires PSM")
	}
	psm := fs.Arg(0)
	endpoints, err := getAPIList(psm)
	if err != nil {
		return 1, err
	}
	if out.JSON {
		return printPrettyJSON(out, endpoints)
	}
	if len(endpoints) == 0 {
		fmt.Fprintf(out.Out, "PSM %s 未找到 API 接口\n", psm)
		return 0, nil
	}
	fmt.Fprintf(out.Out, "PSM: %s\n", psm)
	fmt.Fprintf(out.Out, "接口数量: %d\n\n", len(endpoints))
	for _, ep := range endpoints {
		fmt.Fprintf(out.Out, "  %s\n", ep.Name)
		if ep.Note != "" {
			fmt.Fprintf(out.Out, "    说明:     %s\n", ep.Note)
		}
		fmt.Fprintf(out.Out, "    负责人:   %s\n", ep.Owner)
		fmt.Fprintf(out.Out, "    版本:     %s\n", ep.Version)
		fmt.Fprintf(out.Out, "    序列化:   %s\n", ep.Serializer)
		fmt.Fprintf(out.Out, "    单向调用: %s\n", yesNo(ep.Oneway))
		fmt.Fprintf(out.Out, "    更新时间: %s\n\n", formatUnixSeconds(ep.ModifyTime))
	}
	return 0, nil
}

func psmLinks(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `生成 PSM 的各平台链接

Usage:
  byte-cli psm links [options] <psm-name>

Options:
  -r, --region <region>  区域 (default: "us")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("psm links", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	region := fs.String("region", "us", "区域")
	fs.StringVar(region, "r", "us", "区域")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("links requires PSM")
	}
	links, err := generatePlatformLinks(fs.Arg(0), *region)
	if err != nil {
		return 1, err
	}
	if out.JSON {
		return printPrettyJSON(out, links)
	}
	fmt.Fprintf(out.Out, "PSM: %s\n", links.PSM)
	fmt.Fprintf(out.Out, "区域: %s\n\n", links.Region)
	fmt.Fprintf(out.Out, "TCE:      %s\n", links.TCEURL)
	fmt.Fprintf(out.Out, "SCM:      %s\n", links.SCMURL)
	fmt.Fprintf(out.Out, "TCC:      %s\n", links.TCCURL)
	fmt.Fprintf(out.Out, "Overpass: %s\n", links.OverpassURL)
	return 0, nil
}

func getIdlInfo(psm string) (IDLInfo, error) {
	token, err := tokenFor("cn")
	if err != nil {
		return IDLInfo{}, err
	}
	resp, err := httpclient.Post(idlInfoURL, map[string]any{"PSM": psm}, map[string]string{"X-Jwt-Token": token})
	if err != nil {
		return IDLInfo{}, err
	}
	var payload struct {
		IDLInfo IDLInfo `json:"idl_info"`
	}
	if err := readJSON(resp, &payload); err != nil {
		return IDLInfo{}, err
	}
	if payload.IDLInfo.RepoName == "" && payload.IDLInfo.IDLPath == "" {
		return IDLInfo{}, fmt.Errorf("PSM %s 未找到 IDL 信息", psm)
	}
	payload.IDLInfo.PSM = psm
	if payload.IDLInfo.DefaultBranch == "" {
		payload.IDLInfo.DefaultBranch = "master"
	}
	return payload.IDLInfo, nil
}

func getAPIList(psm string) ([]Endpoint, error) {
	token, err := tokenFor("cn")
	if err != nil {
		return nil, err
	}
	resp, err := httpclient.Get(apiListURL+"?psm="+url.QueryEscape(psm), map[string]string{"X-Jwt-Token": token})
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data []Endpoint `json:"data"`
	}
	if err := readJSON(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Data, nil
}

type PlatformLinks struct {
	PSM         string `json:"psm"`
	Region      string `json:"region"`
	TCEURL      string `json:"tceUrl"`
	SCMURL      string `json:"scmUrl"`
	TCCURL      string `json:"tccUrl"`
	OverpassURL string `json:"overpassUrl"`
}

func generatePlatformLinks(psm, region string) (PlatformLinks, error) {
	domain := psmDomainMap[region]
	if domain == "" {
		return PlatformLinks{}, fmt.Errorf("不支持的区域: %s，可选: us/eu/i18n", region)
	}
	escaped := url.QueryEscape(psm)
	return PlatformLinks{
		PSM:         psm,
		Region:      region,
		TCEURL:      fmt.Sprintf("https://%s/tce/services?keyword=%s&page=1&subs_prefer=true&type=all", domain, psm),
		SCMURL:      fmt.Sprintf("https://%s/scm/favor?page=1&search=%s", domain, escaped),
		TCCURL:      fmt.Sprintf("https://%s/tcc/namespace/%s", domain, psm),
		OverpassURL: fmt.Sprintf("https://cloud.bytedance.net/neptune/overpass/services/%s", psm),
	}, nil
}

func tokenFor(regionName string) (string, error) {
	region, err := auth.ParseRegion(regionName)
	if err != nil {
		return "", err
	}
	return auth.NewManager(region).GetToken(false)
}

func idlRepoURL(info IDLInfo) string {
	if info.RepoName == "" {
		return ""
	}
	return idlRepoBase + "/" + info.RepoName
}

func idlRepoMainURL(info IDLInfo) string {
	if info.RepoName == "" || info.IDLPath == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/blob/%s/%s", idlRepoBase, info.RepoName, info.DefaultBranch, info.IDLPath)
}

func formatUnixSeconds(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006/01/02 15:04")
}

func formatUnixMillis(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.UnixMilli(ts).Format("2006/01/02 15:04")
}

func yesNo(v bool) string {
	if v {
		return "是"
	}
	return "否"
}
