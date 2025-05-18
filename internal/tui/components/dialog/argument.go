package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Argument represents a command argument
type Argument struct {
	Name        string
	Description string
	Required    bool
}

// ArgumentHandler is a function that handles argument values
type ArgumentHandler func(values map[string]string) tea.Cmd