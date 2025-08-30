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

	// Status command styles
	statusHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)   // Blue headers
	statusLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))             // Lighter gray labels (more visible)
	statusValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))             // Light blue values
	statusSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))              // Green for success
	statusWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))             // Yellow for warnings
	statusDangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))             // Red for errors/high values
	statusDivider      = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("─") // Slightly brighter divider
)
