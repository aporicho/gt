package main

import (
	"os"
	"strings"
)

type Theme struct {
	Name    string
	Bg      string
	Fg      string
	Accent  string
	Green   string
	Cyan    string
	Comment string
	Border  string
}

var themes = []Theme{
	{
		Name:    "night",
		Bg:      "#1a1b26",
		Fg:      "#c0caf5",
		Accent:  "#7aa2f7",
		Green:   "#9ece6a",
		Cyan:    "#7dcfff",
		Comment: "#565f89",
		Border:  "#292e42",
	},
	{
		Name:    "storm",
		Bg:      "#24283b",
		Fg:      "#c0caf5",
		Accent:  "#7aa2f7",
		Green:   "#9ece6a",
		Cyan:    "#7dcfff",
		Comment: "#565f89",
		Border:  "#292e42",
	},
	{
		Name:    "moon",
		Bg:      "#222436",
		Fg:      "#c8d3f5",
		Accent:  "#82aaff",
		Green:   "#c3e88d",
		Cyan:    "#86e1fc",
		Comment: "#636da6",
		Border:  "#444a73",
	},
	{
		Name:    "day",
		Bg:      "#e1e2e7",
		Fg:      "#3760bf",
		Accent:  "#2e7de9",
		Green:   "#587539",
		Cyan:    "#118c74",
		Comment: "#848cb5",
		Border:  "#c4c8da",
	},
}

func loadThemeIdx() int {
	data, err := os.ReadFile(themeFile)
	if err != nil {
		return 0
	}
	name := strings.TrimSpace(string(data))
	for i, t := range themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

func saveThemeIdx(idx int) {
	os.WriteFile(themeFile, []byte(themes[idx].Name), 0644)
}
