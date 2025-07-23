//go:build !windows

package cmd

import (
	"path/filepath"
	"runtime"

	"github.com/hyprxlabs/go/env"
)

func homeConfigDir() string {
	dir := env.Get("XDG_CONFIG_HOME")
	if dir == "" {
		home := env.Get("HOME")
		if home == "" {
			user := env.Get("USER")
			if user == "" {
				return ""
			}

			if runtime.GOOS == "darwin" {
				return filepath.Join("/Users", user, "Library", "Application Support", "hyprx", "secrets", "akv")
			}
			return filepath.Join("/home", user, ".config", "hyprx", "secrets", "akv")
		}
		return filepath.Join(home, ".config", "hyprx", "secrets", "akv")
	}

	return filepath.Join(dir, "hyprx", "secrets", "akv")
}

func osConfigDir() string {
	return filepath.Join("/etc", "hyprx", "secrets", "akv")
}
