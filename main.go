package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev" // injected by ldflags at build time

func runApp(startMode appMode) appModel {
	themeIdx := loadThemeIdx()
	m := appModel{
		mode:         startMode,
		themeIdx:     themeIdx,
		cachedStyles: buildStyles(themeIdx),
		dirs:         sortByFrecency(loadBookmarks()),
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

func handleSelected(path string, wantClaude bool) {
	bumpStats(path)
	fmt.Fprint(os.Stderr, "\033[H\033[2J")
	os.WriteFile("/tmp/gt_lastdir", []byte(path), 0644)
	if wantClaude {
		launchClaude(path)
	}
}

func main() {
	args := os.Args[1:]

	binName := filepath.Base(os.Args[0])
	wantClaude := binName == "gtc"

	if len(args) == 0 {
		result := runApp(modePicker)
		if result.selected != "" {
			handleSelected(result.selected, wantClaude)
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
				handleSelected(result.selected, wantClaude)
			}
		}

	case "list":
		for _, d := range loadBookmarks() {
			fmt.Println(d)
		}

	case "update":
		cmd := exec.Command("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/aporicho/gt/main/install.sh | sh")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "更新失败:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("usage:")
		fmt.Println("  gt               pick a bookmark → cd")
		fmt.Println("  gtc              pick a bookmark → cd → launch claude")
		fmt.Println("  gt add           add current directory")
		fmt.Println("  gt add <path>    add given path")
		fmt.Println("  gt list          list all bookmarks")
		fmt.Println("  gt update        update to latest version")
		fmt.Println("  gt --version     print version")
	}
}
