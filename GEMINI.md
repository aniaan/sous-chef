# sous-chef

**sous-chef** is a backend plugin for [mise](https://mise.jdx.dev/), implementing the **vfox** (Version Fox) plugin interface. It serves as a "meta-plugin," allowing `mise` to install and manage versions of various tools (like Neovim, Lazygit, FZF, etc.) through a single plugin backend.

It bridges `mise`'s Lua-based plugin system to a high-performance Go binary that handles the heavy lifting of interacting with GitHub Releases.

## Project Architecture

The project consists of two distinct layers:

### 1. The Lua Interface (vfox/mise Hooks)
These files reside in `hooks/` and define the interface `mise` interacts with. They correspond to the standard `vfox` hooks:

*   **`hooks/backend_list_versions.lua`**: Invoked when listing available versions (e.g., `mise ls-remote`). It delegates to `sous-chef list-versions`.
*   **`hooks/backend_install.lua`**: Invoked to install a specific version (e.g., `mise install neovim@latest`). It delegates to `sous-chef install`.
*   **`hooks/backend_exec_env.lua`**: Defines environment variables for the installed tool. Typically updates `PATH` to include the tool's binary directory.
*   **`lib.lua`**: A bootstrapping helper. It ensures the `sous-chef` Go binary is present on the system (downloading it from GitHub Releases if missing) before any hook attempts to use it.
*   **`metadata.lua`**: Defines plugin metadata (name, version, author).

### 2. The Go Core (Backend Engine)
The Go application (`cmd` / `internal`) is the worker process called by the Lua hooks.

*   **Registry (`internal/registry/registry.go`)**: The central definition file. It contains a hardcoded map of supported tools, defining:
    *   **Repo**: GitHub "owner/repo".
    *   **Asset Patterns**: How to find and name artifacts (e.g., `tool-{{.Version}}-{{.Platform}}.tar.gz`).
    *   **Installation Rules**: Which files to extract and how to handle version string parsing.
*   **Installer (`internal/installer/`)**: Handles downloading, checksum validation (implied or TODO), and extraction.
*   **GitHub Client (`internal/gh/`)**: Interacts with the GitHub API to fetch release tags and assets.

## Usage

### As a mise Plugin

This is the primary usage. To install `sous-chef` as a backend plugin in `mise`:

```bash
mise plugin install sous-chef https://github.com/aniaan/sous-chef
```

Once installed, you can use it to manage tools defined in its registry:

```bash
mise use -g sous-chef:neovim@latest
```

Or via config (only needed if you want to pin the plugin URL in config):

```toml
[plugins]
"vfox-backend:sous-chef" = "https://github.com/aniaan/sous-chef"

[tools]
"sous-chef:neovim" = "latest"
```

Note: the `vfox-backend:` prefix is required only in `[plugins]` config. If the plugin is configured there without the prefix, `mise` may treat it as an asdf plugin and collapse `sous-chef:<tool>` into a single backend, which can break `@latest` resolution across multiple tools. If you previously configured it without the prefix, remove any old installs and reinstall (or delete `~/.local/share/mise/installs/sous-chef-*/.mise.backend`).

### CLI Commands (Go Binary)

The Go binary can be used standalone for debugging or development:

*   **List Versions:** `sous-chef list-versions --tool <name> [--with-published-at]`
*   **Install:** `sous-chef install --tool <name> --version <ver> --dir <path>`
*   **Install Latest:** `sous-chef install-latest --tool <name> --dir <path>`
*   **List Latest (All Tools):** `sous-chef list-latest-versions`

## Development

### Prerequisites

*   **Go**: Version 1.25.5 (as defined in `go.mod`).
*   **Lua**: Useful for testing hooks locally (though `mise` embeds a Lua runtime).

### Building

*   **Build Local:**
    ```bash
    make build
    ```
    Produces a `sous-chef` binary in the root.

*   **Release Builds:**
    ```bash
    make release
    ```
    Builds binaries for Linux/macOS (amd64/arm64).

### Adding a New Tool

To add support for a new tool, modify `internal/registry/registry.go`:

1.  Add a new `PluginConfig` entry to the `Registry` map.
2.  Define:
    *   `Repo`: The GitHub repository.
    *   `AssetTemplate`: Go template for the release filename.
    *   `PlatformMap` / `ArchMap`: Map `sous-chef`'s internal platform/arch constants to the vendor's naming scheme.
3.  Rebuild: `make build`

### Testing

**Testing the Go binary:**
Run the CLI commands directly against the built binary.

**Testing the Lua integration:**
1.  Run `make build` to create a local `sous-chef` binary.
2.  `lib.lua` checks the plugin root for this binary *before* attempting to download one.
3.  You can then trigger `mise` commands that use this plugin context to verify the hooks.

## Key Files

*   `go.mod`: Go dependencies.
*   `metadata.lua`: Plugin definition.
*   `hooks/*.lua`: The entry points called by `mise`.
*   `internal/registry/registry.go`: Tool configuration database.
