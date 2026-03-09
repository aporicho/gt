package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var version = "dev" // injected by ldflags at build time

// ── config paths ──────────────────────────────────────────────────────────────

var (
	bookmarksFile string
	themeFile     string
)

func init() {
	configDir, _ := os.UserConfigDir()
	dir := filepath.Join(configDir, "gtc")
	os.MkdirAll(dir, 0755)
	bookmarksFile = filepath.Join(dir, "bookmarks")
	themeFile = filepath.Join(dir, "theme")
}

// ── bookmarks ─────────────────────────────────────────────────────────────────

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

// ── themes ────────────────────────────────────────────────────────────────────

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

// ── renderer ──────────────────────────────────────────────────────────────────

var renderer = lipgloss.NewRenderer(os.Stderr)

// ── style cache (Fix 3) ───────────────────────────────────────────────────────

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

// ── model ─────────────────────────────────────────────────────────────────────

type model struct {
	dirs          []string
	cursor        int
	selected      string
	width         int
	height        int        // Fix 4: terminal rows
	themeIdx      int
	cachedStyles  themeStyles // Fix 3: pre-built styles
	listOffset    int        // Fix 4: first visible index
	statusMsg     string     // Fix 5: feedback message
	confirmDelete bool       // Fix 6: delete confirmation
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		// clear transient state on every keypress (Fix 5, Fix 6)
		m.statusMsg = ""
		if msg.String() != "-" {
			m.confirmDelete = false
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.dirs)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.dirs) > 0 {
				m.selected = m.dirs[m.cursor]
			}
			return m, tea.Quit
		case "-":
			if len(m.dirs) > 0 {
				if !m.confirmDelete {
					// Fix 6: first press — ask for confirmation
					m.confirmDelete = true
					m.statusMsg = "再按 - 确认删除"
				} else {
					// Fix 6: second press — actually delete
					m.dirs = append(m.dirs[:m.cursor], m.dirs[m.cursor+1:]...)
					saveBookmarks(m.dirs)
					if m.cursor >= len(m.dirs) && m.cursor > 0 {
						m.cursor--
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
					m.cursor = len(m.dirs) - 1
					m.statusMsg = "已添加" // Fix 5: success feedback
				} else {
					m.statusMsg = "已存在" // Fix 5: duplicate feedback
				}
			}
		case "t":
			m.themeIdx = (m.themeIdx + 1) % len(themes)
			saveThemeIdx(m.themeIdx)
			m.cachedStyles = buildStyles(m.themeIdx) // Fix 3: rebuild on theme change
		}
	}

	// Fix 4: clamp scroll offset after every update
	h := m.height
	if h == 0 {
		h = 24
	}
	visibleH := h - 3
	if visibleH < 1 {
		visibleH = 1
	}
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+visibleH {
		m.listOffset = m.cursor - visibleH + 1
	}

	return m, nil
}

func (m model) View() string {
	st := m.cachedStyles

	w := m.width
	if w == 0 {
		w = 60
	}
	// panel has border (1 each side) → panel.Width(w-2) renders at w
	panelW := w - 2

	if len(m.dirs) == 0 {
		content := st.hint.Italic(true).Render("no bookmarks") + "\n" +
			st.hint.Render("use gtc add to add a directory")
		return st.titleBar.Width(w).Render("Bookmarks") + "\n" +
			st.panel.Width(panelW).Render(content) + "\n"
	}

	// title
	header := st.titleBar.Width(w).Render("Bookmarks")

	// Fix 4: compute visible window
	h := m.height
	if h == 0 {
		h = 24
	}
	visibleH := h - 3
	if visibleH < 1 {
		visibleH = 1
	}
	end := m.listOffset + visibleH
	if end > len(m.dirs) {
		end = len(m.dirs)
	}
	visible := m.dirs[m.listOffset:end]

	// list (only visible items)
	var items string
	for i, dir := range visible {
		actualIdx := m.listOffset + i
		base := filepath.Base(dir)
		if actualIdx == m.cursor {
			items += st.cursor.Render("❯ ") + st.selected.Render(base) + "\n"
		} else {
			items += st.normal.Render("  "+base) + "\n"
		}
	}
	panel := st.panel.Width(panelW).Render(strings.TrimRight(items, "\n"))

	// status bar — indented, shorter than panel
	indent := 2
	barW := w - indent*2
	currentPath := m.dirs[m.cursor]

	counter := st.hint.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.dirs)))
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

	// Fix 1: rune-safe path truncation
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

// ── runner ────────────────────────────────────────────────────────────────────

func runPicker(dirs []string) string {
	themeIdx := loadThemeIdx()
	m := model{
		dirs:         dirs,
		themeIdx:     themeIdx,
		cachedStyles: buildStyles(themeIdx), // Fix 3: build styles at creation
	}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return result.(model).selected
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		dirs := loadBookmarks()
		selected := runPicker(dirs)
		if selected != "" {
			if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
				fmt.Fprint(tty, "\033[2J\033[H")
				tty.Close()
			}
			fmt.Println(selected)
		}
		return
	}

	switch args[0] {
	case "--version", "-v":
		fmt.Println("gtc v" + version)
		return

	case "add":
		var dir string
		if len(args) >= 2 {
			dir = args[1]
		} else {
			dir, _ = os.Getwd()
		}
		abs, _ := filepath.Abs(dir)
		dirs := loadBookmarks()
		for _, d := range dirs {
			if d == abs {
				fmt.Println("already exists:", abs)
				return
			}
		}
		dirs = append(dirs, abs)
		saveBookmarks(dirs)
		fmt.Println("added:", abs)

	case "list":
		for _, d := range loadBookmarks() {
			fmt.Println(d)
		}

	default:
		fmt.Println("usage:")
		fmt.Println("  gtc              pick a bookmark")
		fmt.Println("  gtc add          add current directory")
		fmt.Println("  gtc add <path>   add given path")
		fmt.Println("  gtc list         list all bookmarks")
		fmt.Println("  gtc --version    print version")
	}
}
