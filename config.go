package main

import (
	"os"
	"path/filepath"
)

var (
	bookmarksFile string
	themeFile     string
	statsFile     string
)

func init() {
	configDir, _ := os.UserConfigDir()
	dir := filepath.Join(configDir, "gt")
	os.MkdirAll(dir, 0755)
	bookmarksFile = filepath.Join(dir, "bookmarks")
	themeFile = filepath.Join(dir, "theme")
	statsFile = filepath.Join(dir, "stats")
}
