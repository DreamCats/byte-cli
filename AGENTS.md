# Repository Guidelines

## Project Structure & Module Organization

This repository contains `byte-cli`, a TypeScript Node.js CLI. The entrypoint is `src/cli.ts`, which registers Commander subcommands from feature modules under `src/<domain>/`:

- `src/auth/`: cookie configuration, JWT token retrieval, region handling.
- `src/codebase/`: repository and merge request queries.
- `src/logid/`: trace/log lookup and filtering.
- `src/psm/`, `src/iam/`, `src/mcp/`: internal service-specific commands.
- `src/common/`: shared config, cache, HTTP helpers.
- `skills/byte-*`: skill packages installed by `install.sh`.

Build output is generated in `dist/`. Do not edit `dist/` directly.

## Build, Test, and Development Commands

- `npm run dev -- <command>`: run sources with `tsx`, for example `npm run dev -- auth status`.
- `npm run build`: bundle `src/cli.ts` into `dist/cli.js` with `tsup`.
- `npm run typecheck`: run strict TypeScript checking without emitting files.
- `npm link`: expose the built CLI globally as `byte-cli`.
- `bash -n install.sh`: syntax-check the installer after shell changes.

There is currently no `npm test` script or test directory.

## Coding Style & Naming Conventions

Use TypeScript ES modules and keep imports compatible with Node16 resolution; local runtime imports should include `.js` extensions. Follow two-space indentation and Commander-based command structure. Name domain command exports as `<domain>Cmd`, API helpers in `api.ts`, schemas/types in `models.ts`, and shared utilities under `src/common/`.

Prefer Zod for runtime validation and typed interfaces for API responses. Keep output concise and avoid printing secrets; mask cookies and tokens unless a command explicitly returns a token.

## Testing Guidelines

When changing command behavior, run `npm run typecheck` and `npm run build`. For auth, config, installer, or shell changes, also run targeted checks such as `npm run dev -- auth status` or `bash -n install.sh`. Do not claim tests passed unless you ran them.

## Commit & Pull Request Guidelines

Recent commits use Conventional Commit-style prefixes such as `feat:`, `docs:`, and `chore:`. Keep messages short and imperative, for example `feat: add mcp tool lookup`.

Pull requests should describe the changed command or workflow, list verification commands, and call out config or credential-handling impact. Link issues when applicable. Screenshots are usually unnecessary for this CLI.

## Security & Configuration Tips

User configuration is stored in `~/.config/byte-cli/config.yaml`; token cache files are managed under the same config area. Treat cookies, tokens, internal endpoints, and service account secrets as sensitive. Do not add real credentials to docs, examples, tests, or commits.
