package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/DreamCats/byte-cli/internal/auth"
	"github.com/DreamCats/byte-cli/internal/commands"
)

const version = "0.2.0"

type exitError struct {
	code int
	err  error
}

func (e exitError) Error() string { return e.err.Error() }

func Run(args []string) error {
	cfg, rest := parseGlobal(normalizeGlobalFlags(args))
	if len(rest) == 0 || rest[0] == "-h" || rest[0] == "--help" || rest[0] == "help" {
		printTopHelp()
		return nil
	}
	out := commands.NewOutput(cfg.JSON)
	code, err := runCommand(rest[0], rest[1:], out)
	if err != nil {
		if code == 0 {
			code = 1
		}
		return exitError{code: code, err: err}
	}
	if code != 0 {
		return exitError{code: code, err: errors.New("command failed")}
	}
	return nil
}

func Main(args []string) {
	if err := Run(args); err != nil {
		if e, ok := err.(exitError); ok {
			if e.err != nil && e.err.Error() != "command failed" {
				if auth.IsAuthError(e.err) {
					fmt.Fprintf(os.Stderr, "认证错误: %v\n", e.err)
				} else {
					fmt.Fprintf(os.Stderr, "错误: %v\n", e.err)
				}
			}
			os.Exit(e.code)
		}
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

type appConfig struct {
	JSON bool
}

func parseGlobal(args []string) (appConfig, []string) {
	cfg := appConfig{}
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			cfg.JSON = true
		case "--version":
			return cfg, []string{"version"}
		default:
			rest = append(rest, args[i:]...)
			return cfg, rest
		}
	}
	return cfg, rest
}

func normalizeGlobalFlags(args []string) []string {
	if len(args) <= 1 {
		return args
	}
	pulled := []string{}
	rest := []string{}
	for _, arg := range args {
		if arg == "--json" {
			if len(pulled) == 0 {
				pulled = append(pulled, arg)
			}
			continue
		}
		rest = append(rest, arg)
	}
	return append(pulled, rest...)
}

func runCommand(cmd string, args []string, out commands.Output) (int, error) {
	switch cmd {
	case "auth":
		return commands.Auth(args, out)
	case "codebase":
		return commands.Codebase(args, out)
	case "logid":
		return commands.LogID(args, out)
	case "psm":
		return commands.PSM(args, out)
	case "iam":
		return commands.IAM(args, out)
	case "mcp":
		return commands.MCP(args, out)
	case "version":
		fmt.Fprintf(out.Out, "byte-cli %s\n", version)
		return 0, nil
	default:
		return 1, fmt.Errorf("unknown command: %s", cmd)
	}
}

func printTopHelp() {
	fmt.Println(`字节内部开发工具统一 CLI

Usage:
  byte-cli [options] [command]

Options:
  --json        JSON 格式输出
  --version     显示版本信息
  -h, --help    显示帮助信息

Commands:
  auth       认证管理
  codebase   Codebase 仓库和 MR 查询
  logid      通过 Log ID 查询分布式日志链路
  psm        PSM 服务信息查询
  iam        IAM 服务账号查询
  mcp        MCP Server 查询
  version    显示版本信息`)
}
