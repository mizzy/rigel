package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	// Colors
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Border     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Info       lipgloss.Color

	// Styles
	BaseStyle     lipgloss.Style
	HeaderStyle   lipgloss.Style
	InputBoxStyle lipgloss.Style
	ErrorStyle    lipgloss.Style
	SuccessStyle  lipgloss.Style
	WarningStyle  lipgloss.Style
	InfoStyle     lipgloss.Style
	RoleStyle     lipgloss.Style
	TimeStyle     lipgloss.Style
	HelpStyle     lipgloss.Style

	// Sidebar styles
	SidebarStyle      lipgloss.Style
	SidebarTitleStyle lipgloss.Style
	ActiveItemStyle   lipgloss.Style
	InactiveItemStyle lipgloss.Style
	CountStyle        lipgloss.Style

	// Status bar styles
	StatusBarStyle lipgloss.Style
	HelpBarStyle   lipgloss.Style

	// File explorer styles
	FileExplorerStyle lipgloss.Style
	FileStyle         lipgloss.Style
	DirectoryStyle    lipgloss.Style

	// Command palette styles
	CommandPaletteStyle lipgloss.Style
	CommandStyle        lipgloss.Style
	ShortcutStyle       lipgloss.Style

	// Code highlighting styles
	CodeBlockStyle lipgloss.Style
	KeywordStyle   lipgloss.Style
	StringStyle    lipgloss.Style
	CommentStyle   lipgloss.Style
	NumberStyle    lipgloss.Style
	FunctionStyle  lipgloss.Style
}

func NewDefaultTheme() *Theme {
	t := &Theme{
		// Tokyo Night inspired colors
		Primary:    lipgloss.Color("#7aa2f7"),
		Secondary:  lipgloss.Color("#bb9af7"),
		Background: lipgloss.Color("#1a1b26"),
		Foreground: lipgloss.Color("#c0caf5"),
		Border:     lipgloss.Color("#3b4261"),
		Success:    lipgloss.Color("#9ece6a"),
		Warning:    lipgloss.Color("#e0af68"),
		Error:      lipgloss.Color("#f7768e"),
		Info:       lipgloss.Color("#7dcfff"),
	}

	// Base styles
	t.BaseStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(t.Background)

	t.HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	t.InputBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)

	t.ErrorStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true).
		Padding(0, 1)

	t.SuccessStyle = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	t.WarningStyle = lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true)

	t.InfoStyle = lipgloss.NewStyle().
		Foreground(t.Info)

	t.RoleStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.TimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Italic(true)

	t.HelpStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(t.Background).
		Padding(1, 2)

	// Sidebar styles
	t.SidebarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		BorderRight(true).
		Padding(1).
		MarginRight(1)

	t.SidebarTitleStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Align(lipgloss.Center)

	t.ActiveItemStyle = lipgloss.NewStyle().
		Foreground(t.Background).
		Background(t.Primary).
		Bold(true).
		Padding(0, 1)

	t.InactiveItemStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Padding(0, 1)

	t.CountStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89"))

	// Status bar styles
	t.StatusBarStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(lipgloss.Color("#24283b"))

	t.HelpBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Background(t.Background)

	// File explorer styles
	t.FileExplorerStyle = lipgloss.NewStyle().
		Padding(1)

	t.FileStyle = lipgloss.NewStyle().
		Foreground(t.Foreground)

	t.DirectoryStyle = lipgloss.NewStyle().
		Foreground(t.Info).
		Bold(true)

	// Command palette styles
	t.CommandPaletteStyle = lipgloss.NewStyle().
		Padding(1)

	t.CommandStyle = lipgloss.NewStyle().
		Foreground(t.Primary)

	t.ShortcutStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Italic(true)

	// Code highlighting styles
	t.CodeBlockStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#16161e")).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	t.KeywordStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.StringStyle = lipgloss.NewStyle().
		Foreground(t.Success)

	t.CommentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Italic(true)

	t.NumberStyle = lipgloss.NewStyle().
		Foreground(t.Warning)

	t.FunctionStyle = lipgloss.NewStyle().
		Foreground(t.Info)

	return t
}

func NewLightTheme() *Theme {
	t := &Theme{
		// Light theme colors
		Primary:    lipgloss.Color("#0969da"),
		Secondary:  lipgloss.Color("#8250df"),
		Background: lipgloss.Color("#ffffff"),
		Foreground: lipgloss.Color("#1f2328"),
		Border:     lipgloss.Color("#d0d7de"),
		Success:    lipgloss.Color("#1a7f37"),
		Warning:    lipgloss.Color("#9a6700"),
		Error:      lipgloss.Color("#cf222e"),
		Info:       lipgloss.Color("#0969da"),
	}

	// Base styles
	t.BaseStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(t.Background)

	t.HeaderStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	t.InputBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)

	t.ErrorStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true).
		Padding(0, 1)

	t.SuccessStyle = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true)

	t.WarningStyle = lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true)

	t.InfoStyle = lipgloss.NewStyle().
		Foreground(t.Info)

	t.RoleStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.TimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#656d76")).
		Italic(true)

	t.HelpStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(t.Background).
		Padding(1, 2)

	// Sidebar styles
	t.SidebarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		BorderRight(true).
		Padding(1).
		MarginRight(1)

	t.SidebarTitleStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Align(lipgloss.Center)

	t.ActiveItemStyle = lipgloss.NewStyle().
		Foreground(t.Background).
		Background(t.Primary).
		Bold(true).
		Padding(0, 1)

	t.InactiveItemStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Padding(0, 1)

	t.CountStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#656d76"))

	// Status bar styles
	t.StatusBarStyle = lipgloss.NewStyle().
		Foreground(t.Foreground).
		Background(lipgloss.Color("#f6f8fa"))

	t.HelpBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#656d76")).
		Background(t.Background)

	// File explorer styles
	t.FileExplorerStyle = lipgloss.NewStyle().
		Padding(1)

	t.FileStyle = lipgloss.NewStyle().
		Foreground(t.Foreground)

	t.DirectoryStyle = lipgloss.NewStyle().
		Foreground(t.Info).
		Bold(true)

	// Command palette styles
	t.CommandPaletteStyle = lipgloss.NewStyle().
		Padding(1)

	t.CommandStyle = lipgloss.NewStyle().
		Foreground(t.Primary)

	t.ShortcutStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#656d76")).
		Italic(true)

	// Code highlighting styles
	t.CodeBlockStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#f6f8fa")).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	t.KeywordStyle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.StringStyle = lipgloss.NewStyle().
		Foreground(t.Success)

	t.CommentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#656d76")).
		Italic(true)

	t.NumberStyle = lipgloss.NewStyle().
		Foreground(t.Warning)

	t.FunctionStyle = lipgloss.NewStyle().
		Foreground(t.Info)

	return t
}
