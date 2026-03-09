package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	dir := filepath.Join(configDir, "gt")
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

// ── style cache ────────────────────────────────────────────────────────────���─

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

// ── app mode ─────────────────────────────────────────────────────────────────

type appMode int

const (
	modePicker appMode = iota
	modeBrowser
)

// ── unified app model ────────────────────────────────────────────────────────

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

// ── helpers ──────────────────────────────────────────────────────────────────

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
			rmSet := make(map[string]bool, len(m.browsRemoved))
			for p := range m.browsRemoved {
				rmSet[p] = true
			}
			filtered := dirs[:0]
			for _, d := range dirs {
				if !rmSet[d] {
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
	m.dirs = loadBookmarks()
	m.pickerCursor = 0
	m.pickerOffset = 0
	m.statusMsg = ""
	m.confirmDelete = false
	return m
}

// ── update ───────────────────────────────────────────────────────────────────

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

func (m appModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.browsStatus = ""

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m = m.applyBrowserAndSwitchToPicker()
		m.confirmed = false // discard on esc
		// reload without applying
		m.dirs = loadBookmarks()
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

// ── view ─────────────────────────────────────────────────────────────────────

func (m appModel) View() string {
	if m.mode == modeBrowser {
		return m.viewBrowser()
	}
	return m.viewPicker()
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

		var dot string
		if isExisting && isRemoved {
			dot = "○ "
		} else if isExisting {
			dot = "● "
		} else if isSelected {
			dot = "● "
		} else {
			dot = "○ "
		}

		if actualIdx == m.browsCursor {
			if isExisting && !isRemoved {
				items += st.cursor.Render("❯ ") + st.selected.Render(dot) + st.hint.Render(name) + "\n"
			} else if isExisting && isRemoved {
				items += st.cursor.Render("❯ ") + st.normal.Render(dot) + st.hint.Render(name) + "\n"
			} else if isSelected {
				items += st.cursor.Render("❯ ") + st.selected.Render(dot+name) + "\n"
			} else {
				items += st.cursor.Render("❯ ") + st.normal.Render(dot+name) + "\n"
			}
		} else {
			if isExisting && !isRemoved {
				items += "  " + st.selected.Render(dot) + st.hint.Render(name) + "\n"
			} else if isExisting && isRemoved {
				items += "  " + st.normal.Render(dot) + st.hint.Render(name) + "\n"
			} else if isSelected {
				items += "  " + st.selected.Render(dot+name) + "\n"
			} else {
				items += st.normal.Render("  "+dot+name) + "\n"
			}
		}
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

// ── runner ────────────────────────────────────────────────────────────────────

func runApp(startMode appMode) appModel {
	themeIdx := loadThemeIdx()
	m := appModel{
		mode:         startMode,
		themeIdx:     themeIdx,
		cachedStyles: buildStyles(themeIdx),
		dirs:         loadBookmarks(),
	}
	if startMode == modeBrowser {
		m = m.switchToBrowser()
	}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return result.(appModel)
}

// ── launch claude ─────────────────────────────────────────────────────────────

func launchClaude(dir string) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Fprintln(os.Stderr, "未找到 claude 命令，请先安装 Claude Code: https://docs.anthropic.com/en/docs/claude-code")
		os.Exit(1)
	}
	cmd := exec.Command(claudePath)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	args := os.Args[1:]

	binName := filepath.Base(os.Args[0])
	wantClaude := binName == "gtc"

	if len(args) == 0 {
		result := runApp(modePicker)
		if result.selected != "" {
			fmt.Fprint(os.Stderr, "\033[H\033[2J")
			os.WriteFile("/tmp/gt_lastdir", []byte(result.selected), 0644)
			if wantClaude {
				launchClaude(result.selected)
			}
		}
		return
	}

	switch args[0] {
	case "--version", "-v":
		fmt.Println(binName + " v" + version)
		return

	case "add":
		if len(args) >= 2 {
			abs, _ := filepath.Abs(args[1])
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
		} else {
			result := runApp(modeBrowser)
			if result.selected != "" {
				fmt.Fprint(os.Stderr, "\033[H\033[2J")
				os.WriteFile("/tmp/gt_lastdir", []byte(result.selected), 0644)
				if wantClaude {
					launchClaude(result.selected)
				}
			}
		}

	case "list":
		for _, d := range loadBookmarks() {
			fmt.Println(d)
		}

	default:
		fmt.Println("usage:")
		fmt.Println("  gt               pick a bookmark → cd")
		fmt.Println("  gtc              pick a bookmark → cd → launch claude")
		fmt.Println("  gt add           add current directory")
		fmt.Println("  gt add <path>    add given path")
		fmt.Println("  gt list          list all bookmarks")
		fmt.Println("  gt --version     print version")
	}
}
