# doc-helper

CLI tool to concatenate markdown files and sync them to Confluence.

## Repository Structure

```
packages/
├── cli/        # Go CLI — markdown concat & Confluence sync
└── website/    # Astro marketing website
```

## Toolchain

- **CLI:** Go 1.24 (`packages/cli/`)
- **Website:** Astro 6, Tailwind 4, TypeScript (`packages/website/`)

## Commands

```bash
# CLI — build
cd packages/cli && go build -o doc-helper .

# CLI — run
./doc-helper <path> [--exclude <dir>] [--confluence] [--dry-run] [--force]

# CLI — test
cd packages/cli && go test ./...

# Website — dev
cd packages/website && npm run dev

# Website — build
cd packages/website && npm run build
```

## Docs

- [packages/cli/README.md](packages/cli/README.md) — CLI usage, Confluence setup, flags
- [packages/website/README.md](packages/website/README.md) — Website setup

## Guidelines

- After changes, keep these up to date:
    - all README.md files
    - all AGENTS.md files
    - @packages/cli/install.sh
    - the website project
- We want install.sh to work each case:
    - first time installation
    - updating to the latest version
    - no prerequisites installed 
    - prerequisites installed but not in the PATH