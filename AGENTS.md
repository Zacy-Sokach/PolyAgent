# Repository Guidelines

## Project Structure & Module Organization
- `cmd/polyagent/`: CLI entry point (`main.go`) wiring config, TUI, and API clients.
- `internal/api/`, `config/`, `mcp/`, `tui/`, `utils/`, `update/`: core packages for GLM client, settings, MCP tooling, UI, helpers, and self-update logic; tests live next to code (e.g., `internal/config/config_test.go`).
- `scripts/install.sh` and `install.ps1`: installer helpers; `polyagent` is the compiled binary.
- Keep new CLI features inside `internal/tui` or `internal/mcp` with small, focused packages; avoid leaking APIs outside `internal/`.

## Build, Test, and Development Commands
- `go mod download`: fetch dependencies.
- `go run ./cmd/polyagent`: run the app with live code.
- `go build -o polyagent ./cmd/polyagent`: build the release binary.
- `go test ./...`: run all unit tests; add `-run TestName` for targeted debugging.

## Coding Style & Naming Conventions
- Go 1.25+; rely on `gofmt -w .` before committing (tabs for indentation, keep imports grouped).
- Package names are short and lowercase; exported identifiers use PascalCase, locals are camelCase.
- Favor small functions with clear responsibilities; keep I/O and UI side effects near `cmd/` and `internal/tui`.
- Configuration and model constants belong in `internal/config` to centralize defaults and paths.

## Testing Guidelines
- Use the standard library `testing` package; mirror source layout with `_test.go` files.
- Prefer table-driven tests for parsers, config loaders, and tool wiring; stub network calls instead of hitting APIs.
- Ensure new features keep `go test ./...` green; add coverage when touching API/config parsing or user-facing flows.

## Commit & Pull Request Guidelines
- Follow a Conventional Commit style seen in history (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`); keep messages imperative and scoped.
- PRs should include: summary of behavior change, key test commands run, linked issues (if any), and TUI screenshots/gifs when UI output changes.
- Keep diffs small and self-contained; note config or installer impacts explicitly.

## Security & Configuration Tips
- GLM-4.5 API keys are read from `~/.config/polyagent/config.yaml` (or `%APPDATA%\\polyagent\\config.yaml` on Windows); do not commit keys or place them in repo files.
- Validate filesystem operations in `internal/utils` and `internal/mcp` against the current working directory to prevent unwanted writes.
- When adding new external calls, document endpoints and timeouts, and provide opt-in flags or config fields under `internal/config`.
