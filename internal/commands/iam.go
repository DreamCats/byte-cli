package commands

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/DreamCats/byte-cli/internal/httpclient"
)

var iamDomainMap = map[string]string{
	"cn":   "cloud.bytedance.net",
	"us":   "cloud.tiktok-us.net",
	"eu":   "cloud-eu.tiktok-row.net",
	"i18n": "cloud.tiktok-row.net",
}

const (
	iamListPath   = "/api/v1/iam/api/v2/service_account/list_by_page"
	iamSecretPath = "/api/v1/iam/api/v1/service_account/secret"
)

type OwnerInfo struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type SecretItem struct {
	UID       string `json:"uid"`
	Secret    string `json:"secret"`
	Enabled   bool   `json:"enabled"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

type ServiceAccountSpec struct {
	Name          string       `json:"name"`
	NodeID        int          `json:"node_id"`
	Path          string       `json:"path"`
	Description   string       `json:"description"`
	Secret        string       `json:"secret"`
	Owners        []string     `json:"owners"`
	CreatedBy     string       `json:"created_by"`
	UpdatedBy     string       `json:"updated_by"`
	CreatedAt     int64        `json:"created_at"`
	UpdatedAt     int64        `json:"updated_at"`
	I18nName      string       `json:"i18n_name"`
	I18nPath      string       `json:"i18n_path"`
	SensitiveType int          `json:"sensitive_type"`
	OwnerList     []OwnerInfo  `json:"owner_list"`
	Secrets       []SecretItem `json:"secrets"`
}

type ServiceAccount struct {
	Name string             `json:"name"`
	ID   int                `json:"id"`
	Spec ServiceAccountSpec `json:"spec"`
}

type ServiceAccountListResponse struct {
	ErrorCode int              `json:"error_code"`
	Data      []ServiceAccount `json:"data"`
	PageInfo  PageInfo         `json:"page_info"`
}

type PageInfo struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func IAM(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `IAM 服务账号查询

Usage:
  byte-cli iam <command>

Commands:
  list            查询服务账号列表
  secret <name>   查询服务账号密钥`)
		return 0, nil
	}
	switch args[0] {
	case "list":
		return iamList(args[1:], out)
	case "secret":
		return iamSecret(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown iam command: %s", args[0])
	}
}

func iamList(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询服务账号列表

Usage:
  byte-cli iam list [options]

Options:
  -o, --owner <owner>    所有者用户名（默认根据 token 自动识别）
  -r, --region <region>  区域 (default: "cn")
  -p, --page <page>      页码 (default: "1")
  -s, --size <size>      每页数量 (default: "10")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("iam list", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	owner := fs.String("owner", "", "所有者用户名")
	fs.StringVar(owner, "o", "", "所有者用户名")
	region := fs.String("region", "cn", "区域")
	fs.StringVar(region, "r", "cn", "区域")
	page := fs.String("page", "1", "页码")
	fs.StringVar(page, "p", "1", "页码")
	size := fs.String("size", "10", "每页数量")
	fs.StringVar(size, "s", "10", "每页数量")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	pageNum, _ := strconv.Atoi(*page)
	sizeNum, _ := strconv.Atoi(*size)
	resp, err := getServiceAccounts(*owner, *region, pageNum, sizeNum)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, resp)
	}
	if len(resp.Data) == 0 {
		who := "当前用户"
		if *owner != "" {
			who = *owner
		}
		fmt.Fprintf(out.Out, "未找到 %s 的服务账号\n", who)
		return 0, nil
	}
	if *owner != "" {
		fmt.Fprintf(out.Out, "所有者: %s\n", *owner)
	}
	fmt.Fprintf(out.Out, "区域: %s\n", *region)
	fmt.Fprintf(out.Out, "总数: %d\n\n", resp.PageInfo.Total)
	for _, sa := range resp.Data {
		spec := sa.Spec
		fmt.Fprintf(out.Out, "  %s (ID: %d)\n", sa.Name, sa.ID)
		if spec.Description != "" {
			fmt.Fprintf(out.Out, "    描述:     %s\n", spec.Description)
		}
		fmt.Fprintf(out.Out, "    路径:     %s\n", spec.Path)
		fmt.Fprintf(out.Out, "    负责人:   %s\n", strings.Join(spec.Owners, ", "))
		fmt.Fprintf(out.Out, "    创建人:   %s\n", spec.CreatedBy)
		fmt.Fprintf(out.Out, "    创建时间: %s\n", formatUnixMillis(spec.CreatedAt))
		fmt.Fprintf(out.Out, "    更新时间: %s\n\n", formatUnixMillis(spec.UpdatedAt))
	}
	return 0, nil
}

func iamSecret(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `查询服务账号密钥

Usage:
  byte-cli iam secret [options] <name>

Options:
  -r, --region <region>  区域 (default: "cn")
  --json                 JSON 格式输出
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("iam secret", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	region := fs.String("region", "cn", "区域")
	fs.StringVar(region, "r", "cn", "区域")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("secret requires NAME")
	}
	sa, err := getServiceAccountSecret(fs.Arg(0), *region)
	if err != nil {
		return authExitCode(err), err
	}
	if out.JSON {
		return printPrettyJSON(out, sa)
	}
	fmt.Fprintf(out.Out, "服务账号: %s\n", sa.Name)
	fmt.Fprintf(out.Out, "Secret:   %s\n", sa.Spec.Secret)
	if len(sa.Spec.Secrets) > 0 {
		fmt.Fprintf(out.Out, "密钥数量: %d\n", len(sa.Spec.Secrets))
		for _, secret := range sa.Spec.Secrets {
			status := "禁用"
			if secret.Enabled {
				status = "启用"
			}
			fmt.Fprintf(out.Out, "  %s  %s  [%s]\n", secret.UID, secret.Secret, status)
		}
	}
	return 0, nil
}

func getServiceAccounts(owner, region string, page, size int) (ServiceAccountListResponse, error) {
	domain := iamDomainMap[region]
	if domain == "" {
		return ServiceAccountListResponse{}, fmt.Errorf("IAM 不支持区域: %s，可选: cn/us/eu/i18n", region)
	}
	token, err := tokenFor(region)
	if err != nil {
		return ServiceAccountListResponse{}, err
	}
	params := url.Values{"mine_list": {"1"}, "page": {strconv.Itoa(page)}, "size": {strconv.Itoa(size)}}
	if owner != "" {
		params.Set("owner", owner)
		params.Set("owner_check", "1")
	}
	resp, err := httpclient.Get("https://"+domain+iamListPath+"?"+params.Encode(), map[string]string{"X-Jwt-Token": token})
	if err != nil {
		return ServiceAccountListResponse{}, err
	}
	var payload ServiceAccountListResponse
	return payload, readJSON(resp, &payload)
}

func getServiceAccountSecret(name, region string) (ServiceAccount, error) {
	domain := iamDomainMap[region]
	if domain == "" {
		return ServiceAccount{}, fmt.Errorf("IAM 不支持区域: %s，可选: cn/us/eu/i18n", region)
	}
	token, err := tokenFor(region)
	if err != nil {
		return ServiceAccount{}, err
	}
	resp, err := httpclient.Post("https://"+domain+iamSecretPath+"?name="+url.QueryEscape(name), nil, map[string]string{"X-Jwt-Token": token})
	if err != nil {
		return ServiceAccount{}, err
	}
	var payload struct {
		ErrorCode    int            `json:"error_code"`
		ErrorMessage string         `json:"error_message"`
		Data         ServiceAccount `json:"data"`
	}
	if err := readJSON(resp, &payload); err != nil {
		return ServiceAccount{}, err
	}
	if payload.ErrorCode != 0 {
		if payload.ErrorMessage == "" {
			payload.ErrorMessage = "查询密钥失败"
		}
		return ServiceAccount{}, fmt.Errorf("%s", payload.ErrorMessage)
	}
	return payload.Data, nil
}
