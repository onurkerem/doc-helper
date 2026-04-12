# doc-helper

Concatenate all markdown files in a directory and copy the result to your clipboard.

Built for macOS.

## Install

```bash
go install github.com/onurkerem/doc-helper@latest
```

Or with curl:

```bash
curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/install.sh | bash
```

## Usage

```bash
doc-helper <path>
```

Example:

```bash
doc-helper ~/docs/example
```

Output copied to clipboard:

```
<!-- /Users/you/docs/example/1.md -->
content of 1.md

<!-- /Users/you/docs/example/2.md -->
content of 2.md
```

## License

MIT
