# doc-helper

A CLI tool to concatenate markdown files and sync them to Confluence.

## Repository Structure

```
doc-helper/
├── packages/
│   ├── cli/        # Go CLI tool — concatenate & sync markdown to Confluence
│   └── website/    # Marketing website (Astro)
├── LICENSE
└── README.md
```

## Packages

- **[packages/cli](packages/cli/)** — The doc-helper CLI. Concatenates all markdown files in a directory and copies to clipboard, or syncs them as Confluence pages preserving directory hierarchy.
- **[packages/website](packages/website/)** — Marketing website for doc-helper.

## Quick Start

```bash
# Install the CLI
curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/packages/cli/install.sh | bash

# Use it
doc-helper ~/my-docs
```

See [packages/cli/README.md](packages/cli/README.md) for full documentation.

## License

MIT
