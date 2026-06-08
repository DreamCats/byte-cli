# byte-cli

字节内部开发工具统一 CLI（Go 版）。

## Install

One-line install:

```bash
go install github.com/DreamCats/byte-cli/cmd/byte-cli@latest
```

From local source:

```bash
go install ./cmd/byte-cli
```

Development build:

```bash
go build -o byte-cli ./cmd/byte-cli
./byte-cli version
```

## Development

```bash
go test ./...
go build ./...
gofmt -w cmd internal
```

Or use Make:

```bash
make test
make build
make install
```

## Configuration

The Go implementation keeps the TypeScript config layout:

```text
~/.config/byte-cli/
├── config.yaml       # regions.<region>.cookie, proxy.https/http
└── token_cache/      # <region>.json
```

Supported regions:

```text
cn, i18n, us, eu, codebase
```

Cookies and tokens are sensitive. Avoid pasting real credentials into docs,
tests, issues, or commits.

## Common Commands

```bash
byte-cli auth login -r cn
byte-cli auth status
byte-cli auth token -r cn
byte-cli auth config show
byte-cli auth config set-cookie <cookie> -r cn

byte-cli codebase repo info <group/project>
byte-cli codebase mr get <number> -R <group/project>
byte-cli codebase mr comments <number> --unresolved

byte-cli logid <trace-id> -r us -p <psm> -k <keyword>

byte-cli psm idl <psm>
byte-cli psm api-list <psm>
byte-cli psm links <psm> -r us

byte-cli iam list -o <owner> -r cn
byte-cli iam secret <name> -r cn

byte-cli mcp list -s <psm> -r cn
byte-cli mcp tools <server-id>
byte-cli mcp call <server-id> <tool-name> -a key=value

byte-cli metrics tagk <metric-name> -r us
byte-cli metrics tagk <metric-name> --tagkv result=hit -r us
byte-cli metrics field <metric-name> -r us
byte-cli metrics query <metric-name> -r us
byte-cli metrics query <metric-name> --group-by topic -r us
```

Global `--json` is supported, including after subcommands:

```bash
byte-cli --json iam list -r cn
byte-cli iam list -r cn --json
```

## Skills

Agent-facing skills are kept under `skills/byte-*`:

```text
skills/
├── byte-auth/
├── byte-iam/
├── byte-log/
├── byte-mcp/
├── byte-metrics/
├── byte-psm/
└── byte-repo/
```

Install or refresh them with the `skills` CLI from the repository root:

```bash
for skill in skills/byte-*; do
  skills remove "$(basename "$skill")" -g -y 2>/dev/null || true
  skills add "$skill" -g -y
done
```

## Notes

This version preserves the old command surface and config/cache paths while
replacing the TypeScript runtime with a Go binary. Network-backed commands still
depend on valid ByteDance cookies and internal network access.
