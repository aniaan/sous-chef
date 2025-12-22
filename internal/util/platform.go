package util

import (
	"fmt"
	"runtime"
)

// Platform represents the operating system
type Platform string

const (
	Darwin Platform = "darwin"
	Linux  Platform = "linux"
)

// Arch represents the CPU architecture
type Arch string

const (
	X86_64  Arch = "x86_64"
	Aarch64 Arch = "aarch64"
)

// GetSystemInfo returns the current platform and architecture
func GetSystemInfo() (Platform, Arch, error) {
	var p Platform
	switch runtime.GOOS {
	case "darwin":
		p = Darwin
	case "linux":
		p = Linux
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	var a Arch
	switch runtime.GOARCH {
	case "amd64":
		a = X86_64
	case "arm64":
		a = Aarch64
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	return p, a, nil
}
