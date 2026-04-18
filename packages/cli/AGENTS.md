# packages/cli — doc-helper CLI

Go CLI that concatenates markdown files in a directory and either copies to clipboard or syncs to Confluence preserving directory hierarchy.

## Stack

- Go 1.24
- [goldmark](https://github.com/yuin/goldmark) — markdown parsing/conversion

## Commands

```bash
go build -o doc-helper .
go test ./...
go run . <path> [--exclude <dir>] [--confluence] [--dry-run] [--force]
```

## Key Files

- `main.go` — entry point, CLI flags
- `scanner.go` — directory scanning, markdown discovery
- `converter.go` — markdown-to-Confluence storage format conversion
- `sync.go` — Confluence sync orchestration
- `confluence.go` — Confluence REST API client
- `config.go` — config loading from `~/.doc-helper/config.json`
- `state.go` — sync state tracking (content hashes, page IDs)
- `canonical_path.go` — macOS case-insensitive path normalization

## Docs

- [README.md](README.md) — full usage, Confluence setup, examples
