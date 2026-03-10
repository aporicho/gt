package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type appMode int

const (
	modePicker appMode = iota
	modeBrowser
)

type appModel struct {
	mode         appMode
	width        int
	height       int
	themeIdx     int
	cachedStyles themeStyles

	// picker state
	dirs          []string
	pickerCursor  int
	pickerOffset  int
	selected      string // chosen bookmark path
	statusMsg     string
	confirmDelete bool

	// browser state
	currentDir    string
	entries       []os.DirEntry
	browsCursor   int
	browsOffset   int
	browsSelected map[string]bool // paths to add
	browsRemoved  map[string]bool // existing bookmarks to remove
	existingDirs  map[string]bool
	browsStatus   string
	confirmed     bool // browser confirmed with enter
}

func (m appModel) Init() tea.Cmd { return nil }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.mode == modePicker {
			return m.updatePicker(msg)
		}
		return m.updateBrowser(msg)
	}
	return m, nil
}

func (m appModel) View() string {
	if m.mode == modeBrowser {
		return m.viewBrowser()
	}
	return m.viewPicker()
}

func (m appModel) visibleRows(overhead int) int {
	h := m.height
	if h == 0 {
		h = 24
	}
	v := h - overhead
	if v < 1 {
		v = 1
	}
	return v
}
