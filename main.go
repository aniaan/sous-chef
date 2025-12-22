package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/aniaan/sous-chef/internal/gh"
	"github.com/aniaan/sous-chef/internal/installer"
	"github.com/aniaan/sous-chef/internal/registry"
)

var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version":
		fmt.Println(Version)

	case "list-versions":
		listCmd := flag.NewFlagSet("list-versions", flag.ExitOnError)
		tool := listCmd.String("tool", "", "Tool name")
		withPublishedAt := listCmd.Bool("with-published-at", false, "Show published date")
		listCmd.Parse(os.Args[2:])

		if *tool == "" {
			fmt.Println("Error: --tool is required")
			os.Exit(1)
		}
		runListVersions(*tool, *withPublishedAt)

	case "install":
		installCmd := flag.NewFlagSet("install", flag.ExitOnError)
		tool := installCmd.String("tool", "", "Tool name")
		version := installCmd.String("version", "", "Version to install")
		dir := installCmd.String("dir", "", "Installation directory")
		installCmd.Parse(os.Args[2:])

		if *tool == "" || *version == "" || *dir == "" {
			fmt.Println("Error: --tool, --version, and --dir are required")
			os.Exit(1)
		}
		runInstall(*tool, *version, *dir)

	case "install-latest":
		installCmd := flag.NewFlagSet("install-latest", flag.ExitOnError)
		tool := installCmd.String("tool", "", "Tool name")
		dir := installCmd.String("dir", "", "Installation directory")
		installCmd.Parse(os.Args[2:])

		if *tool == "" || *dir == "" {
			fmt.Println("Error: --tool and --dir are required")
			os.Exit(1)
		}
		runInstallLatest(*tool, *dir)

	case "list-latest-versions":
		runListLatestVersions()

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: sous-chef <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  version")
	fmt.Println("  list-versions --tool <name> [--with-published-at]")
	fmt.Println("  list-latest-versions")
	fmt.Println("  install --tool <name> --version <ver> --dir <path>")
	fmt.Println("  install-latest --tool <name> --dir <path>")
}

func runInstallLatest(toolName, dir string) {
	plugin, ok := registry.Registry[toolName]
	if !ok {
		fmt.Printf("Error: Tool '%s' not found in registry\n", toolName)
		os.Exit(1)
	}

	client := gh.NewClient()
	releases, err := plugin.GetReleases(client)
	if err != nil {
		fmt.Printf("Error fetching releases: %v\n", err)
		os.Exit(1)
	}

	if len(releases) == 0 {
		fmt.Printf("Error: No releases found for %s\n", toolName)
		os.Exit(1)
	}

	latest := releases[0]
	displayVersion := plugin.GetDisplayVersion(latest.TagName)

	fmt.Printf("Found latest version: %s (tag: %s)\n", displayVersion, latest.TagName)
	runInstall(toolName, displayVersion, dir)
}

func runListVersions(toolName string, withPublishedAt bool) {
	plugin, ok := registry.Registry[toolName]
	if !ok {
		fmt.Printf("Error: Tool '%s' not found in registry\n", toolName)
		os.Exit(1)
	}

	client := gh.NewClient()
	releases, err := plugin.GetReleases(client)
	if err != nil {
		fmt.Printf("Error fetching releases: %v\n", err)
		os.Exit(1)
	}

	// Limit to top 10
	limit := min(len(releases), 10)

	topReleases := releases[:limit]

	// Print in reverse (oldest first, so newest is at the bottom of the terminal)
	for i := len(topReleases) - 1; i >= 0; i-- {
		r := topReleases[i]
		v := plugin.GetDisplayVersion(r.TagName)

		if withPublishedAt {
			// Format: version #2023-10-27T10:00:00Z
			fmt.Printf("%s #%s\n", v, r.PublishedAt.Format("2006-01-02T15:04:05Z"))
		} else {
			fmt.Println(v)
		}
	}
}

func runListLatestVersions() {
	// Sort plugin names for consistent output
	var plugins []string
	for name := range registry.Registry {
		plugins = append(plugins, name)
	}
	sort.Strings(plugins)

	client := gh.NewClient()

	for _, name := range plugins {
		plugin := registry.Registry[name]
		releases, err := plugin.GetReleases(client)
		if err != nil {
			fmt.Printf("%s: Error fetching releases: %v\n", name, err)
			continue
		}

		if len(releases) == 0 {
			fmt.Printf("%s: No matching releases found\n", name)
			continue
		}

		latest := releases[0]
		v := plugin.GetDisplayVersion(latest.TagName)

		fmt.Printf("%s: %s #%s\n", name, v, latest.PublishedAt.Format("2006-01-02T15:04:05Z"))
	}
}

func runInstall(toolName, version, dir string) {
	plugin, ok := registry.Registry[toolName]
	if !ok {
		fmt.Printf("Error: Tool '%s' not found in registry\n", toolName)
		os.Exit(1)
	}

	err := installer.Install(plugin, version, dir)
	if err != nil {
		fmt.Printf("Error installing %s@%s: %v\n", toolName, version, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully installed %s to %s\n", toolName, dir)
}
