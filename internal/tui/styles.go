package tui

import "github.com/charmbracelet/lipgloss"

// Rigel-inspired color scheme (blue-white star)
var (
	promptSymbol       = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true).Render("✦") // Light blue star symbol
	inputStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))                       // Very light blue-white
	outputStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	thinkingStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Italic(true) // Soft blue
	suggestionStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))              // Gray for suggestions
	highlightStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)    // Highlighted suggestion
	currentModelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)    // Current model in blue
	selectedModelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))              // Selected model in yellow
)
