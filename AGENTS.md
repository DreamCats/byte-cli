# Repository Guidelines

## Project Structure & Module Organization

This repository contains `byte-cli`, a Go CLI migrated from the TypeScript
implementation. The entrypoint is `cmd/byte-cli/main.go`, which delegates to
`internal/cli`.

- `internal/cli/`: top-level argument normalization, command dispatch, exits.
- `internal/commands/`: command implementations for `auth`, `codebase`,
  `logid`, `psm`, `iam`, and `mcp`.
- `internal/auth/`: region definitions, JWT token retrieval, token cache.
- `internal/config/`: `~/.config/byte-cli/config.yaml` loading and saving.
- `internal/httpclient/`: shared HTTP client behavior, timeout, proxy handling.
- `internal/jsonutil/`: API response key normalization helpers.
- `skills/byte-*`: agent-facing skill packages.

Do not add generated build artifacts to the repository. The local `byte-cli`
binary produced by `make build` is disposable.

## Build, Test, and Development Commands

- `go test ./...`: run all unit tests.
- `go build ./...`: verify all packages compile.
- `go build -o byte-cli ./cmd/byte-cli`: build a local CLI binary.
- `go run ./cmd/byte-cli <command>`: run the CLI without installing it.
- `make test`, `make build`, `make install`: wrappers for common Go commands.

When changing command behavior, run `go test ./...` and `go build ./...`.
For help text changes, also spot-check representative commands such as
`go run ./cmd/byte-cli --help` and `go run ./cmd/byte-cli mcp call --help`.

## Coding Style & Naming Conventions

Use idiomatic Go and keep changes small. Run `gofmt` on touched Go files.
Prefer the existing package layout over adding new abstractions. Keep command
parsing close to the relevant command implementation in `internal/commands`.

Use structured JSON decoding for API responses. If an upstream API returns both
PascalCase and snake_case keys, use the existing normalization path instead of
ad hoc string manipulation.

## Testing Guidelines

Add focused tests for parsing, formatting, config handling, and response
normalization. Avoid tests that call internal network APIs. For network-backed
commands, unit-test request-independent behavior and document any manual
verification separately.

Do not claim tests passed unless you ran them in this repository.

## Security & Configuration Tips

User configuration is stored in `~/.config/byte-cli/config.yaml`; token cache
files are stored under `~/.config/byte-cli/token_cache/`. Treat cookies, JWTs,
internal endpoints, and service account secrets as sensitive. Mask credentials
in logs and docs, and never commit real credentials.

Plain output should stay concise. JSON output should preserve useful response
shape without printing secrets unless the command explicitly returns a secret.
