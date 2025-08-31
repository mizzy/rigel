package styles

import "github.com/charmbracelet/lipgloss"

// Status command styles
var (
	StatusHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)   // Blue headers
	StatusLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))             // Lighter gray labels (more visible)
	StatusValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))             // Light blue values
	StatusSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))              // Green for success
	StatusWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))             // Yellow for warnings
	StatusDangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))             // Red for errors/high values
	StatusDivider      = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("─") // Slightly brighter divider
)
