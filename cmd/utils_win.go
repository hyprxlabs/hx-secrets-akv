//go:build windows

package cmd

import (
	"path/filepath"

	"github.com/hyprxlabs/go/env"
)

func homeConfigDir() string {
	dir := env.Get("APPDATA")
	if dir == "" {
		userprofile := env.Get("USERPROFILE")
		if userprofile == "" {
			username := env.Get("USERNAME")
			if username == "" {
				return ""
			}

			return filepath.Join("C:", "Users", username, "AppData", "Roaming", "hyprx", "secrets", "akv")
		}

		if userprofile != "" {
			return filepath.Join(userprofile, "AppData", "Roaming", "hyprx", "secrets", "akv")
		}
	}

	return filepath.Join(dir, "hyprx", "secrets", "akv")
}

func osConfigDir() string {
	dir := env.Get("ALLUSERPROFILE")
	if dir == "" {
		return "C:\\ProgramData"
	}

	return filepath.Join(dir, "hyprx", "secrets", "akv")
}
