package registry

import (
	"sort"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/aniaan/sous-chef/internal/gh"
	"github.com/aniaan/sous-chef/internal/util"
)

// PluginConfig holds the configuration for a tool
type PluginConfig struct {
	Name                    string
	Cmd                     string
	Repo                    string
	AssetTemplate           string // Go template format: bat-v{{.Version}}-{{.Arch}}-{{.Platform}}.tar.gz
	RelativeBinPathTemplate string // Relative path to binary AFTER extraction (and stripping)
	StripComponents         int    // Number of leading directories to strip when extracting
	ReleaseFilter           func(gh.Release) bool
	PlatformMap             map[util.Platform]string
	ArchMap                 map[util.Arch]string
	FormatVersion           func(string) string // GitHub Tag -> Display Version
	RecoverVersion          func(string) string // Display Version -> GitHub Tag
}

// GetReleases fetches, filters, and sorts releases for the plugin
func (p *PluginConfig) GetReleases(client *gh.Client) ([]gh.Release, error) {
	releases, err := client.ListReleases(p.Repo)
	if err != nil {
		return nil, err
	}

	// Filter releases
	var filtered []gh.Release
	for _, r := range releases {
		if p.ReleaseFilter != nil {
			if p.ReleaseFilter(r) {
				filtered = append(filtered, r)
			}
		} else {
			filtered = append(filtered, r)
		}
	}

	// Sort by semantic version (descending)
	// Invalid semver versions are sorted to the end
	sort.Slice(filtered, func(i, j int) bool {
		vi := "v" + p.FormatVersion(filtered[i].TagName)
		vj := "v" + p.FormatVersion(filtered[j].TagName)
		viValid := semver.IsValid(vi)
		vjValid := semver.IsValid(vj)
		if viValid && vjValid {
			return semver.Compare(vi, vj) > 0
		}
		// Valid versions come before invalid ones
		if viValid != vjValid {
			return viValid
		}
		// Both invalid: fall back to PublishedAt
		return filtered[i].PublishedAt.After(filtered[j].PublishedAt)
	})

	return filtered, nil
}

// GetDisplayVersion converts a GitHub tag to a user-friendly version string
func (p *PluginConfig) GetDisplayVersion(tag string) string {
	if p.FormatVersion != nil {
		return p.FormatVersion(tag)
	}
	return tag
}

// Registry stores all known plugins
var Registry = map[string]*PluginConfig{
	"neovim": {
		Name:                    "neovim",
		Cmd:                     "nvim",
		Repo:                    "neovim/neovim",
		AssetTemplate:           "nvim-{{.Platform}}-{{.Arch}}.tar.gz",
		RelativeBinPathTemplate: "bin/nvim",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "macos",
			util.Linux:  "linux",
		},
		ArchMap: map[util.Arch]string{
			util.X86_64:  "x86_64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"rust-analyzer": {
		Name:                    "rust-analyzer",
		Cmd:                     "rust-analyzer",
		Repo:                    "rust-lang/rust-analyzer",
		AssetTemplate:           "rust-analyzer-{{.Arch}}-{{.Platform}}.gz",
		RelativeBinPathTemplate: "rust-analyzer",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-gnu",
		},
		FormatVersion: func(v string) string {
			return strings.ReplaceAll(v, "-", ".")
		},
		RecoverVersion: func(v string) string {
			return strings.ReplaceAll(v, ".", "-")
		},
	},
	"lazygit": {
		Name:                    "lazygit",
		Cmd:                     "lazygit",
		Repo:                    "jesseduffield/lazygit",
		AssetTemplate:           "lazygit_{{.Version}}_{{.Platform}}_{{.Arch}}.tar.gz",
		RelativeBinPathTemplate: "lazygit",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "darwin",
			util.Linux:  "linux",
		},
		ArchMap: map[util.Arch]string{
			util.X86_64:  "x86_64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"fzf": {
		Name:                    "fzf",
		Cmd:                     "fzf",
		Repo:                    "junegunn/fzf",
		AssetTemplate:           "fzf-{{.Version}}-{{.Platform}}_{{.Arch}}.tar.gz",
		RelativeBinPathTemplate: "fzf",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "darwin",
			util.Linux:  "linux",
		},
		ArchMap: map[util.Arch]string{
			util.X86_64:  "amd64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"fd": {
		Name:                    "fd",
		Cmd:                     "fd",
		Repo:                    "sharkdp/fd",
		AssetTemplate:           "fd-v{{.Version}}-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "fd",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-gnu",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"ripgrep": {
		Name:                    "ripgrep",
		Cmd:                     "rg",
		Repo:                    "BurntSushi/ripgrep",
		AssetTemplate:           "ripgrep-{{.Version}}-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "rg",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-musl",
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
	"gh": {
		Name:                    "gh",
		Cmd:                     "gh",
		Repo:                    "cli/cli",
		AssetTemplate:           "gh_{{.Version}}_{{.Platform}}_{{.Arch}}.{{if eq .Platform \"macOS\"}}zip{{else}}tar.gz{{end}}",
		RelativeBinPathTemplate: "bin/gh",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "macOS",
			util.Linux:  "linux",
		},
		ArchMap: map[util.Arch]string{
			util.X86_64:  "amd64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"shfmt": {
		Name:                    "shfmt",
		Cmd:                     "shfmt",
		Repo:                    "mvdan/sh",
		AssetTemplate:           "shfmt_v{{.Version}}_{{.Platform}}_{{.Arch}}",
		RelativeBinPathTemplate: "shfmt",
		StripComponents:         0,
		ArchMap: map[util.Arch]string{
			util.X86_64:  "amd64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"gofumpt": {
		Name:                    "gofumpt",
		Cmd:                     "gofumpt",
		Repo:                    "mvdan/gofumpt",
		AssetTemplate:           "gofumpt_v{{.Version}}_{{.Platform}}_{{.Arch}}",
		RelativeBinPathTemplate: "gofumpt",
		StripComponents:         0,
		ArchMap: map[util.Arch]string{
			util.X86_64:  "amd64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"taplo": {
		Name:                    "taplo",
		Cmd:                     "taplo",
		Repo:                    "tamasfe/taplo",
		AssetTemplate:           "taplo-{{.Platform}}-{{.Arch}}.gz",
		RelativeBinPathTemplate: "taplo",
		StripComponents:         0,
		ReleaseFilter: func(r gh.Release) bool {
			// Only accept simple version tags like "0.10.0"
			if len(r.TagName) == 0 {
				return false
			}
			return r.TagName[0] >= '0' && r.TagName[0] <= '9'
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
	"stylua": {
		Name:                    "stylua",
		Cmd:                     "stylua",
		Repo:                    "JohnnyMorganz/StyLua",
		AssetTemplate:           "stylua-{{.Platform}}-{{.Arch}}.zip",
		RelativeBinPathTemplate: "stylua",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "macos",
			util.Linux:  "linux",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"lua-language-server": {
		Name:                    "lua-language-server",
		Cmd:                     "lua-language-server",
		Repo:                    "LuaLS/lua-language-server",
		AssetTemplate:           "lua-language-server-{{.Version}}-{{.Platform}}-{{.Arch}}.tar.gz",
		RelativeBinPathTemplate: "bin/lua-language-server",
		StripComponents:         0,
		ArchMap: map[util.Arch]string{
			util.X86_64:  "x64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
	"starship": {
		Name:                    "starship",
		Cmd:                     "starship",
		Repo:                    "starship/starship",
		AssetTemplate:           "starship-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "starship",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-gnu",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"zoxide": {
		Name:                    "zoxide",
		Cmd:                     "zoxide",
		Repo:                    "ajeetdsouza/zoxide",
		AssetTemplate:           "zoxide-{{.Version}}-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "zoxide",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-musl",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"uv": {
		Name:                    "uv",
		Cmd:                     "uv",
		Repo:                    "astral-sh/uv",
		AssetTemplate:           "uv-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "uv",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-gnu",
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
	"tree-sitter": {
		Name:                    "tree-sitter",
		Cmd:                     "tree-sitter",
		Repo:                    "tree-sitter/tree-sitter",
		AssetTemplate:           "tree-sitter-{{.Platform}}-{{.Arch}}.gz",
		RelativeBinPathTemplate: "tree-sitter",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "macos",
			util.Linux:  "linux",
		},
		ArchMap: map[util.Arch]string{
			util.X86_64:  "x64",
			util.Aarch64: "arm64",
		},
		FormatVersion:  RemoveVPrefixFormatVersion,
		RecoverVersion: AddVPrefixRecoverVersion,
	},
	"ty": {
		Name:                    "ty",
		Cmd:                     "ty",
		Repo:                    "astral-sh/ty",
		AssetTemplate:           "ty-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "ty",
		StripComponents:         1,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-gnu",
		},
		ReleaseFilter: func(r gh.Release) bool {
			return !r.Prerelease
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
	"codex": {
		Name:                    "codex",
		Cmd:                     "codex",
		Repo:                    "openai/codex",
		AssetTemplate:           "codex-{{.Arch}}-{{.Platform}}.tar.gz",
		RelativeBinPathTemplate: "codex-{{.Arch}}-{{.Platform}}",
		StripComponents:         0,
		PlatformMap: map[util.Platform]string{
			util.Darwin: "apple-darwin",
			util.Linux:  "unknown-linux-musl",
		},
		ReleaseFilter: func(r gh.Release) bool {
			return strings.HasPrefix(r.TagName, "rust-v") && !r.Prerelease
		},
		RecoverVersion: func(v string) string {
			if strings.HasPrefix(v, "rust-v") {
				return v
			}
			return "rust-v" + strings.TrimPrefix(v, "v")
		},
		FormatVersion: func(v string) string {
			return strings.TrimPrefix(v, "rust-v")
		},
	},
	"zls": {
		Name:                    "zls",
		Cmd:                     "zls",
		Repo:                    "zigtools/zls",
		AssetTemplate:           "zls-{{.Arch}}-{{.Platform}}.tar.xz",
		RelativeBinPathTemplate: "zls",
		StripComponents:         0,
		ArchMap: map[util.Arch]string{
			util.X86_64:  "x86_64",
			util.Aarch64: "aarch64",
		},
		PlatformMap: map[util.Platform]string{
			util.Darwin: "macos",
			util.Linux:  "linux",
		},
		FormatVersion:  NoOpVersion,
		RecoverVersion: NoOpVersion,
	},
}

func NoOpVersion(v string) string {
	return v
}

func RemoveVPrefixFormatVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// DefaultRecoverVersion adds 'v' prefix
func AddVPrefixRecoverVersion(v string) string {
	if !strings.HasPrefix(v, "v") && len(v) > 0 && v[0] >= '0' && v[0] <= '9' {
		return "v" + v
	}
	return v
}
