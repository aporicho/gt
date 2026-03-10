package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var renderer = lipgloss.NewRenderer(os.Stderr)

type themeStyles struct {
	titleBar lipgloss.Style
	selected lipgloss.Style
	cursor   lipgloss.Style
	normal   lipgloss.Style
	hint     lipgloss.Style
	panel    lipgloss.Style
}

func buildStyles(themeIdx int) themeStyles {
	th := themes[themeIdx]
	c := func(hex string) lipgloss.Color { return lipgloss.Color(hex) }
	return themeStyles{
		titleBar: renderer.NewStyle().
			Bold(true).
			Foreground(c(th.Bg)).
			Background(c(th.Accent)).
			Align(lipgloss.Center),
		selected: renderer.NewStyle().Foreground(c(th.Green)).Bold(true),
		cursor:   renderer.NewStyle().Foreground(c(th.Cyan)).Bold(true),
		normal:   renderer.NewStyle().Foreground(c(th.Fg)),
		hint:     renderer.NewStyle().Foreground(c(th.Comment)),
		panel: renderer.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c(th.Border)).
			Padding(0, 1),
	}
}
