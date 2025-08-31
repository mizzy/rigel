package tui

import "github.com/charmbracelet/lipgloss"

// Rigel-inspired color scheme (blue-white star)
// Note: Most styles moved to render package for better organization
var (

	// Status command styles
	statusHeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true)   // Blue headers
	statusLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))             // Lighter gray labels (more visible)
	statusValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("195"))             // Light blue values
	statusSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))              // Green for success
	statusWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))             // Yellow for warnings
	statusDangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))             // Red for errors/high values
	statusDivider      = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("─") // Slightly brighter divider
)
