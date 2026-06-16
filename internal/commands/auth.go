package commands

import (
	"flag"
	"fmt"

	"github.com/DreamCats/byte-cli/internal/auth"
	"github.com/DreamCats/byte-cli/internal/config"
)

func Auth(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printAuthHelp(out)
		return 0, nil
	}
	switch args[0] {
	case "login":
		return authLogin(args[1:], out)
	case "status":
		return authStatus(out)
	case "token":
		return authToken(args[1:], out)
	case "config":
		return authConfig(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown auth command: %s", args[0])
	}
}

func authLogin(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `获取并缓存 JWT Token。云区域未配置可用 Cookie 时会启动浏览器登录

Usage:
  byte-cli auth login [options]

Options:
  -r, --region <region>  区域: cn/i18n/us/eu/codebase (default: "cn")
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("auth login", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	regionOpt := fs.String("region", "cn", "区域: cn/i18n/us/eu/codebase")
	fs.StringVar(regionOpt, "r", "cn", "区域: cn/i18n/us/eu/codebase")
	if err := fs.Parse(normalizeFlags(args, stringSet("r", "region"), nil)); err != nil {
		return 1, err
	}
	region, err := auth.ParseRegion(*regionOpt)
	if err != nil {
		return 1, err
	}
	token, err := auth.NewManager(region).LoginInteractive(out.Err)
	if err != nil {
		return authExitCode(err), err
	}
	preview := token
	if len(preview) > 20 {
		preview = preview[:20] + "..."
	}
	fmt.Fprintf(out.Out, "区域 %s 认证成功\n", region.Value)
	fmt.Fprintf(out.Out, "Token: %s\n", preview)
	return 0, nil
}

func authStatus(out Output) (int, error) {
	cfg, err := config.Load()
	if err != nil {
		return 1, err
	}
	for _, region := range auth.AllRegions() {
		hasCookie := config.GetCookie(cfg, region.Value) != ""
		tokenValid := hasCookie && auth.IsTokenValid(region.Value)
		cookieStatus := "✗"
		tokenStatus := "✗"
		if hasCookie {
			cookieStatus = "✓"
		}
		if tokenValid {
			tokenStatus = "✓"
		}
		fmt.Fprintf(out.Out, "  %10s  Cookie: %s  Token: %s\n", region.Value, cookieStatus, tokenStatus)
	}
	return 0, nil
}

func authToken(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `输出指定区域的 JWT Token

Usage:
  byte-cli auth token [options]

Options:
  -r, --region <region>  区域: cn/i18n/us/eu/codebase (default: "cn")
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("auth token", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	regionOpt := fs.String("region", "cn", "区域: cn/i18n/us/eu/codebase")
	fs.StringVar(regionOpt, "r", "cn", "区域: cn/i18n/us/eu/codebase")
	if err := fs.Parse(normalizeFlags(args, stringSet("r", "region"), nil)); err != nil {
		return 1, err
	}
	region, err := auth.ParseRegion(*regionOpt)
	if err != nil {
		return 1, err
	}
	token, err := auth.NewManager(region).GetToken(false)
	if err != nil {
		return authExitCode(err), err
	}
	fmt.Fprintln(out.Out, token)
	return 0, nil
}

func authConfig(args []string, out Output) (int, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		printHelp(out, `配置管理

Usage:
  byte-cli auth config <command>

Commands:
  show                  查看当前配置
  set-cookie <cookie>   设置指定区域的 Cookie`)
		return 0, nil
	}
	switch args[0] {
	case "show":
		return authConfigShow(out)
	case "set-cookie":
		return authConfigSetCookie(args[1:], out)
	default:
		return 1, fmt.Errorf("unknown auth config command: %s", args[0])
	}
}

func authConfigShow(out Output) (int, error) {
	cfg, err := config.Load()
	if err != nil {
		return 1, err
	}
	fmt.Fprintln(out.Out, "区域 Cookie:")
	for _, region := range auth.AllRegions() {
		cookie := config.GetCookie(cfg, region.Value)
		if cookie == "" {
			fmt.Fprintf(out.Out, "  %s: (未配置)\n", region.Value)
			continue
		}
		fmt.Fprintf(out.Out, "  %s: %s\n", region.Value, config.MaskCookie(cookie))
	}
	if cfg.Proxy.HTTPS != "" || cfg.Proxy.HTTP != "" {
		fmt.Fprintln(out.Out, "\n代理:")
		if cfg.Proxy.HTTPS != "" {
			fmt.Fprintf(out.Out, "  HTTPS: %s\n", cfg.Proxy.HTTPS)
		}
		if cfg.Proxy.HTTP != "" {
			fmt.Fprintf(out.Out, "  HTTP: %s\n", cfg.Proxy.HTTP)
		}
	}
	return 0, nil
}

func authConfigSetCookie(args []string, out Output) (int, error) {
	if wantsHelp(args) {
		printHelp(out, `设置指定区域的 Cookie

Usage:
  byte-cli auth config set-cookie [options] <cookie>

Options:
  -r, --region <region>  区域: cn/i18n/us/eu/codebase (default: "cn")
  -h, --help             显示帮助信息`)
		return 0, nil
	}
	fs := flag.NewFlagSet("auth config set-cookie", flag.ContinueOnError)
	fs.SetOutput(out.Err)
	regionOpt := fs.String("region", "cn", "区域: cn/i18n/us/eu/codebase")
	fs.StringVar(regionOpt, "r", "cn", "区域: cn/i18n/us/eu/codebase")
	if err := fs.Parse(normalizeFlags(args, stringSet("r", "region"), nil)); err != nil {
		return 1, err
	}
	if fs.NArg() == 0 {
		return 1, fmt.Errorf("set-cookie requires COOKIE")
	}
	region, err := auth.ParseRegion(*regionOpt)
	if err != nil {
		return 1, err
	}
	cfg, err := config.Load()
	if err != nil {
		return 1, err
	}
	config.SetCookie(&cfg, region.Value, fs.Arg(0))
	if err := config.Save(cfg); err != nil {
		return 1, err
	}
	fmt.Fprintf(out.Out, "区域 %s Cookie 已更新\n", region.Value)
	return 0, nil
}

func authExitCode(err error) int {
	if auth.IsAuthError(err) {
		return 2
	}
	return 1
}

func printAuthHelp(out Output) {
	printHelp(out, `认证管理

Usage:
  byte-cli auth <command>

Commands:
  login                  获取并缓存 JWT Token
  status                 查看各区域认证状态
  token                  输出指定区域的 JWT Token
  config                 配置管理`)
}
