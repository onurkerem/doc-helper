# doc-helper

Concatenate all markdown files in a directory and copy the result to your clipboard.

Built for macOS.

## Install / Update

```bash
curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/install.sh | bash
```

## Usage

```bash
doc-helper <path> [--exclude <dir>[,<dir>...]]
```

Output copied to clipboard:

```
<!-- /Users/you/docs/example/1.md -->
content of 1.md

<!-- /Users/you/docs/example/2.md -->
content of 2.md
```

### Exclude directories

```bash
# Single directory
doc-helper <path> --exclude node_modules

# Multiple directories
doc-helper <path> --exclude node_modules --exclude .git

# Comma-separated
doc-helper <path> --exclude node_modules,.git,vendor
```

## License

MIT
