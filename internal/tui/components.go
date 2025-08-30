package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Sidebar component
type Sidebar struct {
	theme  *Theme
	items  []SidebarItem
	width  int
	active int
}

type SidebarItem struct {
	Icon  string
	Title string
	Count int
}

func NewSidebar(theme *Theme) *Sidebar {
	return &Sidebar{
		theme: theme,
		width: 20,
		items: []SidebarItem{
			{Icon: "💬", Title: "Chat", Count: 0},
			{Icon: "📁", Title: "Files", Count: 0},
			{Icon: "📝", Title: "History", Count: 0},
			{Icon: "⚙️", Title: "Settings", Count: 0},
		},
		active: 0,
	}
}

func (s *Sidebar) Width() int {
	return s.width
}

func (s *Sidebar) Render(height int) string {
	var items []string

	// Title
	title := s.theme.SidebarTitleStyle.Render("✨ Rigel AI")
	items = append(items, title)
	items = append(items, strings.Repeat("─", s.width-2))

	// Menu items
	for i, item := range s.items {
		var itemStr string
		if i == s.active {
			itemStr = s.theme.ActiveItemStyle.Render(
				fmt.Sprintf("%s %s", item.Icon, item.Title),
			)
		} else {
			itemStr = s.theme.InactiveItemStyle.Render(
				fmt.Sprintf("%s %s", item.Icon, item.Title),
			)
		}

		if item.Count > 0 {
			count := s.theme.CountStyle.Render(fmt.Sprintf("(%d)", item.Count))
			itemStr = fmt.Sprintf("%s %s", itemStr, count)
		}

		items = append(items, itemStr)
	}

	// Join all items
	content := strings.Join(items, "\n")

	// Apply sidebar style and set height
	style := s.theme.SidebarStyle.
		Width(s.width).
		Height(height)

	return style.Render(content)
}

// StatusBar component
type StatusBar struct {
	theme  *Theme
	height int
}

func NewStatusBar(theme *Theme) *StatusBar {
	return &StatusBar{
		theme:  theme,
		height: 3,
	}
}

func (s *StatusBar) Height() int {
	return s.height
}

func (s *StatusBar) Render(width int, state sessionState, messageCount int) string {
	// Get working directory
	wd, _ := os.Getwd()
	dir := filepath.Base(wd)

	// State indicator
	var stateStr string
	switch state {
	case stateChat:
		stateStr = "CHAT"
	case stateFileExplorer:
		stateStr = "FILES"
	case stateCommandPalette:
		stateStr = "COMMAND"
	case stateHelp:
		stateStr = "HELP"
	}

	// Build status items
	left := fmt.Sprintf(" 📂 %s", dir)
	center := fmt.Sprintf("💬 Messages: %d", messageCount)
	right := fmt.Sprintf("[%s] ", stateStr)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	centerWidth := lipgloss.Width(center)
	rightWidth := lipgloss.Width(right)

	spaces := width - leftWidth - centerWidth - rightWidth - 2
	if spaces < 0 {
		spaces = 0
	}

	leftSpaces := spaces / 2
	rightSpaces := spaces - leftSpaces

	// Build status line
	statusLine := left + strings.Repeat(" ", leftSpaces) + center + strings.Repeat(" ", rightSpaces) + right

	// Help line
	helpLine := " Ctrl+H: Help | Ctrl+E: Files | Ctrl+P: Commands | Ctrl+Q: Quit"

	// Combine lines
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		s.theme.StatusBarStyle.Width(width).Render(statusLine),
		s.theme.HelpBarStyle.Width(width).Render(helpLine),
	)

	return content
}

// FileExplorer component
type FileExplorer struct {
	theme *Theme
	list  list.Model
}

type FileItem struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

func (i FileItem) FilterValue() string { return i.Name }
func (i FileItem) Title() string {
	if i.IsDir {
		return fmt.Sprintf("📁 %s", i.Name)
	}
	return fmt.Sprintf("📄 %s", i.Name)
}
func (i FileItem) Description() string {
	if i.IsDir {
		return "Directory"
	}
	return formatFileSize(i.Size)
}

func NewFileExplorer(theme *Theme) *FileExplorer {
	items := []list.Item{}

	// Load current directory files
	entries, _ := os.ReadDir(".")
	for _, entry := range entries {
		info, _ := entry.Info()
		item := FileItem{
			Name:  entry.Name(),
			Path:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		items = append(items, item)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "File Explorer"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &FileExplorer{
		theme: theme,
		list:  l,
	}
}

func (f *FileExplorer) Init() tea.Cmd {
	return nil
}

func (f *FileExplorer) Update(msg tea.Msg) (*FileExplorer, tea.Cmd) {
	var cmd tea.Cmd
	f.list, cmd = f.list.Update(msg)
	return f, cmd
}

func (f *FileExplorer) Render(width, height int) string {
	f.list.SetSize(width, height)
	return f.theme.FileExplorerStyle.Render(f.list.View())
}

// CommandPalette component
type CommandPalette struct {
	theme *Theme
	list  list.Model
}

type CommandItem struct {
	Name     string
	Desc     string
	Shortcut string
	Action   func() tea.Cmd
}

func (i CommandItem) FilterValue() string { return i.Name }
func (i CommandItem) Title() string       { return i.Name }
func (i CommandItem) Description() string {
	if i.Shortcut != "" {
		return fmt.Sprintf("%s (%s)", i.Desc, i.Shortcut)
	}
	return i.Desc
}

func NewCommandPalette(theme *Theme) *CommandPalette {
	items := []list.Item{
		CommandItem{
			Name:     "Clear Chat",
			Desc:     "Clear the current chat session",
			Shortcut: "Ctrl+L",
		},
		CommandItem{
			Name:     "Save Session",
			Desc:     "Save the current session to a file",
			Shortcut: "Ctrl+S",
		},
		CommandItem{
			Name: "Load Session",
			Desc: "Load a previous session from file",
		},
		CommandItem{
			Name: "Export Markdown",
			Desc: "Export chat as markdown",
		},
		CommandItem{
			Name: "Change Theme",
			Desc: "Switch between light and dark themes",
		},
		CommandItem{
			Name: "Run Command",
			Desc: "Execute a shell command",
		},
		CommandItem{
			Name: "Git Status",
			Desc: "Show git repository status",
		},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Command Palette"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &CommandPalette{
		theme: theme,
		list:  l,
	}
}

func (c *CommandPalette) Init() tea.Cmd {
	return nil
}

func (c *CommandPalette) Update(msg tea.Msg) (*CommandPalette, tea.Cmd) {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

func (c *CommandPalette) Render(width, height int) string {
	c.list.SetSize(width, height)
	return c.theme.CommandPaletteStyle.Render(c.list.View())
}

// Helper functions
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
