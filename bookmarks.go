package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadBookmarks() []string {
	f, err := os.Open(bookmarksFile)
	if err != nil {
		return nil
	}
	defer f.Close()
	var dirs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			dirs = append(dirs, line)
		}
	}
	return dirs
}

func saveBookmarks(dirs []string) {
	f, err := os.Create(bookmarksFile)
	if err != nil {
		return
	}
	defer f.Close()
	for _, d := range dirs {
		fmt.Fprintln(f, d)
	}
}
