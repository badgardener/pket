package ui

import "charm.land/lipgloss/v2"

var (
	Success  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	Error    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	Warning  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	Info     = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	Muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	Title    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true)
	Command  = lipgloss.NewStyle().Foreground(lipgloss.Color("#94E2D5")).Bold(true)
	Flag     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387"))
	Argument = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5C2E7"))
	Package  = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
)
