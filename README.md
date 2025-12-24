# sous-chef

sous-chef is a mise backend (vfox plugin) that lets you install many tools through a single plugin.

## Quick start

Install the plugin:

```bash
mise plugin install sous-chef https://github.com/aniaan/sous-chef
```

Install a tool:

```bash
mise use -g sous-chef:neovim@latest
```

## Configure in mise.toml

Pin the plugin (optional):

```toml
[plugins]
"vfox-backend:sous-chef" = "https://github.com/aniaan/sous-chef"
```

Use tools:

```toml
[tools]
"sous-chef:neovim" = "latest"
"sous-chef:lazygit" = "latest"
```

Note: the `vfox-backend:` prefix is required in `[plugins]`. Without it, mise may treat sous-chef as an asdf plugin and collapse tool names, which breaks `@latest` across multiple tools.

## Supported tools

- neovim
- rust-analyzer
- lazygit
- fzf
- fd
- ripgrep
- gh
- shfmt
- gofumpt
- taplo
- stylua
- lua-language-server
- starship
- zoxide
- uv
- tree-sitter
- ty
- codex

See `internal/registry/registry.go` for full details and asset patterns.

## GitHub API rate limits

sous-chef queries GitHub Releases. Unauthenticated requests are limited to 60/hour. To avoid rate limits, set a token:

```bash
export GITHUB_TOKEN="your_token_here"
```

## CLI (for debugging)

The Go binary can be used directly:

```bash
sous-chef list-versions --tool <name>
sous-chef install --tool <name> --version <ver> --dir <path>
sous-chef install-latest --tool <name> --dir <path>
sous-chef list-latest-versions
```

## Development

Build:

```bash
make build
```

Add a new tool:

1. Add a new entry to `internal/registry/registry.go`.
2. Define repo and asset template maps.
3. Rebuild with `make build`.

## Links

- mise: https://mise.jdx.dev/
