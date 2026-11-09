# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based developer toolkit called `dev` that provides various utilities for software developers. It's a command-line application built with Cobra that offers tools for search, file operations, encoding/decoding, time conversion, JSON/SQL processing, and more.

## Development Commands

### Building and Running
```bash
# Build the application
go build -ldflags="-s -w" -o dev ./main.go

# Run the application
go run main.go

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Code Quality
```bash
# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Run pre-commit hooks
pre-commit run --all-files
```

### Docker
```bash
# Build Docker image
docker build -t dev .

# Run Docker container
docker run -d -p 8899:8899 --name dev_test dev
```

## Architecture

### Directory Structure
- `cmd/` - Command implementations using Cobra CLI framework
- `internal/` - Internal business logic packages
  - `search/` - Search engine integration and web scraping
  - `json2struct/` - JSON to Go struct conversion
  - `sql2struct/` - SQL to Go struct conversion (MySQL focused)
  - `timer/` - Time and timestamp utilities
  - `word/` - Word case conversion utilities
  - `tools/` - General utility functions
  - `updater/` - Self-update support: version resolution, update check, release download with sha256 verify and atomic swap, go install delegation

### Key Components

**Command Structure (cmd/)**
- `root.go` - Main command initialization and configuration management
- Each command file implements a specific feature (search, encode, decode, etc.)
- Commands use Cobra for CLI argument parsing and help generation

**Configuration System**
- Uses Viper for configuration management
- Default config file: `$HOME/.dev.yaml`
- Supports configuration for HTTP server, search defaults, and SQL connections
- Configuration priority: command flags > config file > defaults

**Search Engine Integration**
- Supports multiple search engines (Bing, Baidu, Google, GitHub, etc.)
- Two modes: browser opening and terminal display
- Web scraping using goquery for HTML parsing
- AJAX API support for dynamic content

**Interactive Console**
- Provides an interactive shell-like interface
- Uses go-prompt for autocompletion
- Supports common developer utilities (encoding, search, file operations)

## Configuration

The application uses a YAML configuration file at `$HOME/.dev.yaml` with these sections:

```yaml
http:
    port: 8899
search:
    cli_is_desc: true
    default_engine: bing
    default_type: cli
sql:
    type: mysql
    host: 127.0.0.1:3306
    username: root
    password: ""
    db: ""
    charset: utf8mb4
update:
    api_base: https://api.github.com  # GitHub API endpoint override (default api.github.com), for proxy/mirror setups
```

## Dependencies

Key third-party libraries:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/PuerkitoBio/goquery` - HTML parsing
- `github.com/c-bata/go-prompt` - Interactive prompts
- `github.com/pterm/pterm` - Terminal styling

## Testing

The project uses Go's built-in testing framework. Tests are located alongside the source code and can be run with `go test ./...`. Coverage reports can be generated with the commands listed above.

## Version Management

Version is resolved at startup via `internal/updater` with three-level priority: ldflags injection (`-ldflags "-X github.com/wujunwei928/dev/cmd.Version=X.Y.Z"`) > `debug.ReadBuildInfo` main module version (go install builds) > `devel+<short commit>` (local builds). Do not hardcode version constants in `cmd/root.go`.

## Pre-commit Hooks

The project uses pre-commit hooks for code quality:
- YAML/JSON/TOML validation
- Byte order marker checks
- Go unit tests
- Large file detection

## Agent skills

### Issue tracker

Issues 追踪于 GitHub Issues（wujunwei928/dev），全部操作走 `gh` CLI。见 `docs/agents/issue-tracker.md`。

### Triage labels

采用默认五标签：`needs-triage` / `needs-info` / `ready-for-agent` / `ready-for-human` / `wontfix`。见 `docs/agents/triage-labels.md`。

### Domain docs

单上下文布局：根 `GLOSSARY.md`（惰性创建）+ `docs/adr/`。见 `docs/agents/domain.md`。