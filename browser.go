package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func loadEntries(dir string) []os.DirEntry {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []os.DirEntry
	for _, e := range items {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() < dirs[j].Name()
	})
	return dirs
}

func (m appModel) switchToBrowser() appModel {
	cwd, _ := os.Getwd()
	existing := make(map[string]bool, len(m.dirs))
	for _, d := range m.dirs {
		existing[d] = true
	}
	m.mode = modeBrowser
	m.currentDir = cwd
	m.entries = loadEntries(cwd)
	m.browsCursor = 0
	m.browsOffset = 0
	m.browsSelected = make(map[string]bool)
	m.browsRemoved = make(map[string]bool)
	m.existingDirs = existing
	m.browsStatus = ""
	m.confirmed = false
	return m
}

func (m appModel) applyBrowserAndSwitchToPicker() appModel {
	if m.confirmed {
		dirs := loadBookmarks()
		// remove
		if len(m.browsRemoved) > 0 {
			filtered := dirs[:0]
			for _, d := range dirs {
				if !m.browsRemoved[d] {
					filtered = append(filtered, d)
				}
			}
			dirs = filtered
		}
		// add
		added := make([]string, 0, len(m.browsSelected))
		for p := range m.browsSelected {
			added = append(added, p)
		}
		sort.Strings(added)
		dirs = append(dirs, added...)
		saveBookmarks(dirs)
	}
	// switch back to picker
	m.mode = modePicker
	m.dirs = sortByFrecency(loadBookmarks())
	m.pickerCursor = 0
	m.pickerOffset = 0
	m.statusMsg = ""
	m.confirmDelete = false
	return m
}

func (m appModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.browsStatus = ""

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m = m.applyBrowserAndSwitchToPicker()
		m.confirmed = false // discard on esc
		// reload without applying
		m.dirs = sortByFrecency(loadBookmarks())
		return m, nil
	case "up", "k":
		if m.browsCursor > 0 {
			m.browsCursor--
		}
	case "down", "j":
		if m.browsCursor < len(m.entries)-1 {
			m.browsCursor++
		}
	case "right":
		if len(m.entries) > 0 {
			sub := filepath.Join(m.currentDir, m.entries[m.browsCursor].Name())
			m.currentDir = sub
			m.entries = loadEntries(sub)
			m.browsCursor = 0
			m.browsOffset = 0
		}
	case "left":
		parent := filepath.Dir(m.currentDir)
		if parent != m.currentDir {
			oldBase := filepath.Base(m.currentDir)
			m.currentDir = parent
			m.entries = loadEntries(parent)
			m.browsCursor = 0
			for i, e := range m.entries {
				if e.Name() == oldBase {
					m.browsCursor = i
					break
				}
			}
			m.browsOffset = 0
		}
	case " ", "tab":
		if len(m.entries) > 0 {
			abs := filepath.Join(m.currentDir, m.entries[m.browsCursor].Name())
			if m.existingDirs[abs] {
				if m.browsRemoved[abs] {
					delete(m.browsRemoved, abs)
				} else {
					m.browsRemoved[abs] = true
				}
			} else if m.browsSelected[abs] {
				delete(m.browsSelected, abs)
			} else {
				m.browsSelected[abs] = true
			}
		}
	case "enter":
		if len(m.browsSelected) > 0 || len(m.browsRemoved) > 0 {
			m.confirmed = true
			m = m.applyBrowserAndSwitchToPicker()
			return m, nil
		}
		m.browsStatus = "no changes"
	case "t":
		m.themeIdx = (m.themeIdx + 1) % len(themes)
		saveThemeIdx(m.themeIdx)
		m.cachedStyles = buildStyles(m.themeIdx)
	}

	// clamp scroll
	visibleH := m.visibleRows(4)
	if m.browsCursor < m.browsOffset {
		m.browsOffset = m.browsCursor
	}
	if m.browsCursor >= m.browsOffset+visibleH {
		m.browsOffset = m.browsCursor - visibleH + 1
	}
	return m, nil
}

func (m appModel) viewBrowser() string {
	st := m.cachedStyles
	w := m.width
	if w == 0 {
		w = 60
	}
	panelW := w - 2

	header := st.titleBar.Width(w).Render("添加书签")
	pathLine := st.hint.Render("  " + m.currentDir)

	if len(m.entries) == 0 {
		content := st.hint.Italic(true).Render("（空目录）")
		return header + "\n" + pathLine + "\n" +
			st.panel.Width(panelW).Render(content) + "\n"
	}

	visibleH := m.visibleRows(4)
	end := m.browsOffset + visibleH
	if end > len(m.entries) {
		end = len(m.entries)
	}
	visible := m.entries[m.browsOffset:end]

	var items string
	for i, entry := range visible {
		actualIdx := m.browsOffset + i
		abs := filepath.Join(m.currentDir, entry.Name())
		name := entry.Name() + "/"

		isExisting := m.existingDirs[abs]
		isSelected := m.browsSelected[abs]
		isRemoved := m.browsRemoved[abs]

		var prefix string
		if actualIdx == m.browsCursor {
			prefix = st.cursor.Render("❯ ")
		} else {
			prefix = "  "
		}

		var dot string
		if (isExisting && !isRemoved) || isSelected {
			dot = "● "
		} else {
			dot = "○ "
		}

		var line string
		switch {
		case isExisting && !isRemoved:
			line = st.selected.Render(dot) + st.hint.Render(name)
		case isExisting && isRemoved:
			line = st.normal.Render(dot) + st.hint.Render(name)
		case isSelected:
			line = st.selected.Render(dot + name)
		default:
			line = st.normal.Render(dot + name)
		}
		items += prefix + line + "\n"
	}
	panel := st.panel.Width(panelW).Render(strings.TrimRight(items, "\n"))

	// status bar
	indent := 2
	barW := w - indent*2

	addCount := len(m.browsSelected)
	rmCount := len(m.browsRemoved)
	var countParts []string
	if addCount > 0 {
		countParts = append(countParts, fmt.Sprintf("+%d", addCount))
	}
	if rmCount > 0 {
		countParts = append(countParts, fmt.Sprintf("-%d", rmCount))
	}
	var counterStr string
	if len(countParts) > 0 {
		counterStr = strings.Join(countParts, " ")
	}
	counter := st.hint.Render(counterStr)

	var hints string
	if m.browsStatus != "" {
		hints = st.hint.Render(m.browsStatus)
	} else {
		hints = st.hint.Render("space:select  enter:confirm  t:theme  esc:back")
	}
	gap := barW - lipgloss.Width(hints) - lipgloss.Width(counter)
	if gap < 0 {
		gap = 0
	}
	statusBar := strings.Repeat(" ", indent) +
		counter +
		strings.Repeat(" ", gap) +
		hints

	return header + "\n" + pathLine + "\n" + panel + "\n" + statusBar + "\n"
}
