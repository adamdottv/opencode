package config

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMapConfig defines the configuration for keyboard shortcuts
type KeyMapConfig struct {
	// Chat keymaps
	NewSession           []string `json:"newSession,omitempty"`
	Cancel               []string `json:"cancel,omitempty"`
	ToggleTools          []string `json:"toggleTools,omitempty"`
	ShowCompletionDialog []string `json:"showCompletionDialog,omitempty"`

	// Global keymaps
	ViewLogs      []string `json:"viewLogs,omitempty"`
	Quit          []string `json:"quit,omitempty"`
	Help          []string `json:"help,omitempty"`
	SwitchSession []string `json:"switchSession,omitempty"`
	Commands      []string `json:"commands,omitempty"`
	FilePicker    []string `json:"filePicker,omitempty"`
	Models        []string `json:"models,omitempty"`
	Theme         []string `json:"theme,omitempty"`
	Tools         []string `json:"tools,omitempty"`
}

// DefaultKeyMapConfig returns the default keyboard shortcut configuration
func DefaultKeyMapConfig() *KeyMapConfig {
	return &KeyMapConfig{
		// Chat keymaps
		NewSession:           []string{"ctrl+n"},
		Cancel:               []string{"esc"},
		ToggleTools:          []string{"ctrl+h"},
		ShowCompletionDialog: []string{"/"},

		// Global keymaps
		ViewLogs:      []string{"ctrl+l"},
		Quit:          []string{"ctrl+c"},
		Help:          []string{"ctrl+_"},
		SwitchSession: []string{"ctrl+s"},
		Commands:      []string{"ctrl+k"},
		FilePicker:    []string{"ctrl+f"},
		Models:        []string{"ctrl+o"},
		Theme:         []string{"ctrl+t"},
		Tools:         []string{"f9"},
	}
}

func (c *Config) GetAllKeyBinding() []key.Binding {
	chatKeyMap := c.GetChatKeyMap()
	globalKeyMap := c.GetGlobalKeyMap()

	return []key.Binding{
		chatKeyMap.NewSession,
		chatKeyMap.Cancel,
		chatKeyMap.ToggleTools,
		chatKeyMap.ShowCompletionDialog,
		globalKeyMap.ViewLogs,
		globalKeyMap.Quit,
		globalKeyMap.Help,
		globalKeyMap.SwitchSession,
		globalKeyMap.Commands,
		globalKeyMap.FilePicker,
		globalKeyMap.Models,
		globalKeyMap.Theme,
		globalKeyMap.Tools,
	}
}

// GetChatKeyMap returns a ChatKeyMap with bindings from config
func (c *Config) GetChatKeyMap() ChatKeyMap {
	keys := c.KeyMaps

	out := ChatKeyMap{
		NewSession: key.NewBinding(
			key.WithKeys(keys.NewSession...),
			key.WithHelp(keys.NewSession[0], "new session"),
		),
		Cancel: key.NewBinding(
			key.WithKeys(keys.Cancel...),
			key.WithHelp(keys.Cancel[0], "cancel"),
		),
		ToggleTools: key.NewBinding(
			key.WithKeys(keys.ToggleTools...),
			key.WithHelp(keys.ToggleTools[0], "toggle tools"),
		),
		ShowCompletionDialog: key.NewBinding(
			key.WithKeys(keys.ShowCompletionDialog...),
			key.WithHelp(keys.ShowCompletionDialog[0], "complete"),
		),
	}

	return out
}

// ChatKeyMap defines key bindings for the chat page
type ChatKeyMap struct {
	NewSession           key.Binding
	Cancel               key.Binding
	ToggleTools          key.Binding
	ShowCompletionDialog key.Binding
}

// GlobalKeyMap defines key bindings for global application functions
type GlobalKeyMap struct {
	ViewLogs      key.Binding
	Quit          key.Binding
	Help          key.Binding
	SwitchSession key.Binding
	Commands      key.Binding
	FilePicker    key.Binding
	Models        key.Binding
	Theme         key.Binding
	Tools         key.Binding
}

// GetGlobalKeyMap returns a GlobalKeyMap with bindings from config
func (c *Config) GetGlobalKeyMap() GlobalKeyMap {
	keys := c.KeyMaps

	return GlobalKeyMap{
		ViewLogs: key.NewBinding(
			key.WithKeys(keys.ViewLogs...),
			key.WithHelp(keys.ViewLogs[0], "logs"),
		),
		Quit: key.NewBinding(
			key.WithKeys(keys.Quit...),
			key.WithHelp(keys.Quit[0], "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(keys.Help...),
			key.WithHelp("ctrl+?", "toggle help"),
		),
		SwitchSession: key.NewBinding(
			key.WithKeys(keys.SwitchSession...),
			key.WithHelp(keys.SwitchSession[0], "switch session"),
		),
		Commands: key.NewBinding(
			key.WithKeys(keys.Commands...),
			key.WithHelp(keys.Commands[0], "commands"),
		),
		FilePicker: key.NewBinding(
			key.WithKeys(keys.FilePicker...),
			key.WithHelp(keys.FilePicker[0], "select files to upload"),
		),
		Models: key.NewBinding(
			key.WithKeys(keys.Models...),
			key.WithHelp(keys.Models[0], "model selection"),
		),
		Theme: key.NewBinding(
			key.WithKeys(keys.Theme...),
			key.WithHelp(keys.Theme[0], "switch theme"),
		),
		Tools: key.NewBinding(
			key.WithKeys(keys.Tools...),
			key.WithHelp(keys.Tools[0], "show available tools"),
		),
	}
}

// mergeKeyMaps merges user-provided keymaps with default keymaps
// If a keymap is not provided by the user, the default is used
func mergeKeyMaps(userKeyMaps, defaultKeyMaps *KeyMapConfig) {
	// Chat keymaps
	if userKeyMaps.NewSession == nil {
		userKeyMaps.NewSession = defaultKeyMaps.NewSession
	}
	if userKeyMaps.Cancel == nil {
		userKeyMaps.Cancel = defaultKeyMaps.Cancel
	}
	if userKeyMaps.ToggleTools == nil {
		userKeyMaps.ToggleTools = defaultKeyMaps.ToggleTools
	}
	if userKeyMaps.ShowCompletionDialog == nil {
		userKeyMaps.ShowCompletionDialog = defaultKeyMaps.ShowCompletionDialog
	}

	// Global keymaps
	if userKeyMaps.ViewLogs == nil {
		userKeyMaps.ViewLogs = defaultKeyMaps.ViewLogs
	}
	if userKeyMaps.Quit == nil {
		userKeyMaps.Quit = defaultKeyMaps.Quit
	}
	if userKeyMaps.Help == nil {
		userKeyMaps.Help = defaultKeyMaps.Help
	}
	if userKeyMaps.SwitchSession == nil {
		userKeyMaps.SwitchSession = defaultKeyMaps.SwitchSession
	}
	if userKeyMaps.Commands == nil {
		userKeyMaps.Commands = defaultKeyMaps.Commands
	}
	if userKeyMaps.FilePicker == nil {
		userKeyMaps.FilePicker = defaultKeyMaps.FilePicker
	}
	if userKeyMaps.Models == nil {
		userKeyMaps.Models = defaultKeyMaps.Models
	}
	if userKeyMaps.Theme == nil {
		userKeyMaps.Theme = defaultKeyMaps.Theme
	}
	if userKeyMaps.Tools == nil {
		userKeyMaps.Tools = defaultKeyMaps.Tools
	}
}
