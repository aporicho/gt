package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m appModel) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.statusMsg = ""
	if msg.String() != "-" {
		m.confirmDelete = false
	}

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
	case "down", "j":
		if m.pickerCursor < len(m.dirs)-1 {
			m.pickerCursor++
		}
	case "enter":
		if len(m.dirs) == 0 {
			m = m.switchToBrowser()
			return m, nil
		}
		m.selected = m.dirs[m.pickerCursor]
		return m, tea.Quit
	case "-":
		if len(m.dirs) > 0 {
			if !m.confirmDelete {
				m.confirmDelete = true
				m.statusMsg = "再按 - 确认删除"
			} else {
				m.dirs = append(m.dirs[:m.pickerCursor], m.dirs[m.pickerCursor+1:]...)
				saveBookmarks(m.dirs)
				if m.pickerCursor >= len(m.dirs) && m.pickerCursor > 0 {
					m.pickerCursor--
				}
				m.confirmDelete = false
			}
		}
	case "+":
		if cwd, err := os.Getwd(); err == nil {
			exists := false
			for _, d := range m.dirs {
				if d == cwd {
					exists = true
					break
				}
			}
			if !exists {
				m.dirs = append(m.dirs, cwd)
				saveBookmarks(m.dirs)
				m.pickerCursor = len(m.dirs) - 1
				m.statusMsg = "已添加"
			} else {
				m.statusMsg = "已存在"
			}
		}
	case "t":
		m.themeIdx = (m.themeIdx + 1) % len(themes)
		saveThemeIdx(m.themeIdx)
		m.cachedStyles = buildStyles(m.themeIdx)
	}

	// clamp scroll
	visibleH := m.visibleRows(3)
	if m.pickerCursor < m.pickerOffset {
		m.pickerOffset = m.pickerCursor
	}
	if m.pickerCursor >= m.pickerOffset+visibleH {
		m.pickerOffset = m.pickerCursor - visibleH + 1
	}
	return m, nil
}

func (m appModel) viewPicker() string {
	st := m.cachedStyles
	w := m.width
	if w == 0 {
		w = 60
	}
	panelW := w - 2

	if len(m.dirs) == 0 {
		content := st.hint.Italic(true).Render("还没有书签") + "\n" +
			st.hint.Render("按 Enter 进入目录浏览器添加书签")
		return st.titleBar.Width(w).Render("Bookmarks") + "\n" +
			st.panel.Width(panelW).Render(content) + "\n"
	}

	header := st.titleBar.Width(w).Render("Bookmarks")

	visibleH := m.visibleRows(3)
	end := m.pickerOffset + visibleH
	if end > len(m.dirs) {
		end = len(m.dirs)
	}
	visible := m.dirs[m.pickerOffset:end]

	var items string
	for i, dir := range visible {
		actualIdx := m.pickerOffset + i
		base := filepath.Base(dir)
		if actualIdx == m.pickerCursor {
			items += st.cursor.Render("❯ ") + st.selected.Render(base) + "\n"
		} else {
			items += st.normal.Render("  "+base) + "\n"
		}
	}
	panel := st.panel.Width(panelW).Render(strings.TrimRight(items, "\n"))

	indent := 2
	barW := w - indent*2
	currentPath := m.dirs[m.pickerCursor]

	counter := st.hint.Render(fmt.Sprintf("%d/%d", m.pickerCursor+1, len(m.dirs)))
	var hints string
	if m.statusMsg != "" {
		hints = st.hint.Render(m.statusMsg)
	} else {
		hints = st.hint.Render("↵:open  -:del  +:add  t:theme")
	}

	hintW := lipgloss.Width(hints)
	countW := lipgloss.Width(counter)
	pathMaxW := barW - hintW - countW - 2
	if pathMaxW < 0 {
		pathMaxW = 0
	}

	runes := []rune(currentPath)
	if len(runes) > pathMaxW {
		runes = append([]rune("…"), runes[len(runes)-pathMaxW+1:]...)
		currentPath = string(runes)
	}

	pathR := st.hint.Render(currentPath)
	gap := barW - lipgloss.Width(pathR) - hintW - countW
	if gap < 0 {
		gap = 0
	}
	statusBar := strings.Repeat(" ", indent) +
		pathR +
		strings.Repeat(" ", gap/2) + hints +
		strings.Repeat(" ", gap-gap/2) + counter

	return header + "\n" + panel + "\n" + statusBar + "\n"
}
