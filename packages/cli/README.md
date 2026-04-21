# doc-helper

Concatenate all markdown files in a directory and copy the result to your clipboard.

Optionally sync markdown content to Confluence, creating a matching page hierarchy.

Built for macOS.

## Install / Update

```bash
curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/packages/cli/install.sh | bash
```

## Usage

```bash
doc-helper <path> [--exclude <dir>[,<dir>...]] [--confluence] [--dry-run] [--force]
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

### Confluence Sync

Use `--confluence` to sync your markdown files as Confluence pages alongside the clipboard copy.

```bash
doc-helper ~/my-docs --confluence
doc-helper ~/my-docs --confluence --dry-run   # preview without creating pages
doc-helper ~/my-docs --confluence --force       # re-upload all pages (ignore content hash)
```

`--force` only applies with `--confluence`. Use it after changing doc-helper's conversion logic (for example) so every page is pushed again even when the markdown files are unchanged.

Given this directory structure:

```
~/my-docs/
  overview.md
  getting-started/
    installation.md
    configuration.md
  guides/
    api-usage.md
```

And a configured parent page "Documentation", doc-helper creates:

```
Documentation (existing, not modified)
├── Overview                (content from overview.md)
├── getting-started         (empty container page)
│   ├── Installation        (content from installation.md)
│   └── Configuration       (content from configuration.md)
└── guides                  (empty container page)
    └── API Usage           (content from api-usage.md)
```

Page titles come from the first `# Heading` in each file (falls back to the filename with dashes/underscores replaced by spaces, title-cased). Directories become empty container pages.

#### Setup

1. Create the config directory:

   ```bash
   mkdir -p ~/.doc-helper
   ```

2. Create `~/.doc-helper/config.json`:

   ```json
   {
     "syncs": [
       {
         "path": "/Users/you/my-docs",
         "confluence_base_url": "https://yourcompany.atlassian.net/wiki",
         "email": "you@company.com",
         "api_token": "YOUR_API_TOKEN",
         "parent_page_id": "123456",
         "exclude_files": ["CLAUDE.md", "AGENTS.md"]
       }
     ]
   }
   ```

3. Get your Confluence API token:
   - Go to https://id.atlassian.com/manage-profile/security/api-tokens
   - Create a token and paste it in the config

4. Find the parent page ID from the page URL:
   - `https://yourcompany.atlassian.net/wiki/spaces/PROJ/pages/123456/Page+Title`
   - The page ID is `123456`

#### Exclude files

Use `exclude_files` in config to skip specific file names during Confluence sync. Matching is by exact file name (case-sensitive) at any directory depth.

```json
"exclude_files": ["CLAUDE.md", "AGENTS.md"]
```

#### How it works

- Sync roots are resolved to the directory names stored on disk (macOS APFS is typically case-insensitive but case-preserving, so `documents` and `Documents` in a path match the same folder). The `path` in `config.json` does not need to match letter case exactly, and CLI arguments are normalized the same way.
- On first run, pages are created under the configured parent page
- On subsequent runs, only changed files are updated (tracked by content hash)
- State is stored in `~/.doc-helper/state.json` (auto-managed)
- If state is lost, existing pages are detected by title to avoid duplicates. Before every update, the live Confluence page is fetched so the correct version number is used (cached state can be stale after edits in Confluence or interrupted runs; list endpoints do not always include version metadata)
- `--dry-run` shows what would happen without making API calls
- `--confluence --force` skips the "unchanged file" hash check and updates every markdown-backed page
- Fenced code blocks in markdown (GitHub-style triple-backtick fences with an optional language) are published as Confluence **Code Block** macros (native code snippets: syntax highlighting and copy). Inline backtick code is left as normal formatted text.

## License

MIT
