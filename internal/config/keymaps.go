// Package config manages application configuration from various sources.
package config

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMapConfig defines the configuration for keyboard shortcuts
type KeyMapConfig struct {
	Chat     *ChatKeyMapConfig     `json:"chat,omitempty"`
	Global   *GlobalKeyMapConfig   `json:"global,omitempty"`
	Editor   *EditorKeyMapConfig   `json:"editor,omitempty"`
	Messages *MessagesKeyMapConfig `json:"messages,omitempty"`
	Logs     *LogsKeyMapConfig     `json:"logs,omitempty"`
}

// ChatKeyMapConfig defines keyboard shortcuts for the chat page
type ChatKeyMapConfig struct {
	NewSession           []string `json:"newSession,omitempty"`
	Cancel               []string `json:"cancel,omitempty"`
	ToggleTools          []string `json:"toggleTools,omitempty"`
	ShowCompletionDialog []string `json:"showCompletionDialog,omitempty"`
}

// GlobalKeyMapConfig defines keyboard shortcuts for global application functions
type GlobalKeyMapConfig struct {
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

// EditorKeyMapConfig defines keyboard shortcuts for the text editor
type EditorKeyMapConfig struct {
	Submit []string `json:"submit,omitempty"`
	Clear  []string `json:"clear,omitempty"`
}

// MessagesKeyMapConfig defines keyboard shortcuts for message navigation
type MessagesKeyMapConfig struct {
	HalfPageUp   []string `json:"halfPageUp,omitempty"`
	HalfPageDown []string `json:"halfPageDown,omitempty"`
}

// LogsKeyMapConfig defines keyboard shortcuts for logs page navigation
type LogsKeyMapConfig struct {
	Back []string `json:"back,omitempty"`
}

// DefaultKeyMapConfig returns the default keyboard shortcut configuration
func DefaultKeyMapConfig() *KeyMapConfig {
	return &KeyMapConfig{
		Chat: &ChatKeyMapConfig{
			NewSession:           []string{"ctrl+n"},
			Cancel:               []string{"esc"},
			ToggleTools:          []string{"ctrl+h"},
			ShowCompletionDialog: []string{"/"},
		},
		Global: &GlobalKeyMapConfig{
			ViewLogs:      []string{"ctrl+l"},
			Quit:          []string{"ctrl+c"},
			Help:          []string{"ctrl+_"},
			SwitchSession: []string{"ctrl+s"},
			Commands:      []string{"ctrl+k"},
			FilePicker:    []string{"ctrl+f"},
			Models:        []string{"ctrl+o"},
			Theme:         []string{"ctrl+t"},
			Tools:         []string{"f9"},
		},
		Editor: &EditorKeyMapConfig{
			Submit: []string{"ctrl+j", "ctrl+enter"},
			Clear:  []string{"ctrl+u"},
		},
		Messages: &MessagesKeyMapConfig{
			HalfPageUp:   []string{"ctrl+u"},
			HalfPageDown: []string{"ctrl+d"},
		},
		Logs: &LogsKeyMapConfig{
			Back: []string{"esc"},
		},
	}
}

func (c *Config) GetAllKeyBinding() []key.Binding {
	chatKeyMap := c.GetChatKeyMap()
	globalKeyMap := c.GetGlobalKeyMap()
	editorKeyMap := c.GetEditorKeyMap()
	messagesKeyMap := c.GetMessagesKeyMap()
	logsKeyMap := c.GetLogsKeyMap()

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
		editorKeyMap.Submit,
		editorKeyMap.Clear,
		messagesKeyMap.HalfPageUp,
		messagesKeyMap.HalfPageDown,
		logsKeyMap.Back,
	}
}

// GetChatKeyMap returns a ChatKeyMap with bindings from config
func (c *Config) GetChatKeyMap() ChatKeyMap {
	keys := c.KeyMaps.Chat

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
	keys := c.KeyMaps.Global

	return GlobalKeyMap{
		ViewLogs: key.NewBinding(
			key.WithKeys(keys.ViewLogs...),
			key.WithHelp(keys.ViewLogs[0], "view logs"),
		),
		Quit: key.NewBinding(
			key.WithKeys(keys.Quit...),
			key.WithHelp(keys.Quit[0], "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(keys.Help...),
			key.WithHelp("ctrl+?", "help"), // Special case for display
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
			key.WithHelp(keys.FilePicker[0], "file picker"),
		),
		Models: key.NewBinding(
			key.WithKeys(keys.Models...),
			key.WithHelp(keys.Models[0], "models"),
		),
		Theme: key.NewBinding(
			key.WithKeys(keys.Theme...),
			key.WithHelp(keys.Theme[0], "theme"),
		),
		Tools: key.NewBinding(
			key.WithKeys(keys.Tools...),
			key.WithHelp(keys.Tools[0], "tools"),
		),
	}
}

// EditorKeyMap defines key bindings for the text editor
type EditorKeyMap struct {
	Submit key.Binding
	Clear  key.Binding
}

// GetEditorKeyMap returns an EditorKeyMap with bindings from config
func (c *Config) GetEditorKeyMap() EditorKeyMap {
	keys := c.KeyMaps.Editor

	return EditorKeyMap{
		Submit: key.NewBinding(
			key.WithKeys(keys.Submit...),
			key.WithHelp(keys.Submit[0], "submit"),
		),
		Clear: key.NewBinding(
			key.WithKeys(keys.Clear...),
			key.WithHelp(keys.Clear[0], "clear"),
		),
	}
}

// MessagesKeyMap defines key bindings for message navigation
type MessagesKeyMap struct {
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
}

// GetMessagesKeyMap returns a MessagesKeyMap with bindings from config
func (c *Config) GetMessagesKeyMap() MessagesKeyMap {
	keys := c.KeyMaps.Messages

	return MessagesKeyMap{
		HalfPageUp: key.NewBinding(
			key.WithKeys(keys.HalfPageUp...),
			key.WithHelp(keys.HalfPageUp[0], "half page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys(keys.HalfPageDown...),
			key.WithHelp(keys.HalfPageDown[0], "half page down"),
		),
	}
}

// LogsKeyMap defines key bindings for logs page navigation
type LogsKeyMap struct {
	Back key.Binding
}

// GetLogsKeyMap returns a LogsKeyMap with bindings from config
func (c *Config) GetLogsKeyMap() LogsKeyMap {
	keys := c.KeyMaps.Logs

	return LogsKeyMap{
		Back: key.NewBinding(
			key.WithKeys(keys.Back...),
			key.WithHelp(keys.Back[0], "back"),
		),
	}
}

// mergeKeyMaps merges user-provided keymaps with default keymaps
// If a keymap is not provided by the user, the default is used
func mergeKeyMaps(userKeyMaps, defaultKeyMaps *KeyMapConfig) {
	// Merge Chat keymaps
	if userKeyMaps.Chat == nil {
		userKeyMaps.Chat = defaultKeyMaps.Chat
	} else {
		if userKeyMaps.Chat.NewSession == nil {
			userKeyMaps.Chat.NewSession = defaultKeyMaps.Chat.NewSession
		}
		if userKeyMaps.Chat.Cancel == nil {
			userKeyMaps.Chat.Cancel = defaultKeyMaps.Chat.Cancel
		}
		if userKeyMaps.Chat.ToggleTools == nil {
			userKeyMaps.Chat.ToggleTools = defaultKeyMaps.Chat.ToggleTools
		}
		if userKeyMaps.Chat.ShowCompletionDialog == nil {
			userKeyMaps.Chat.ShowCompletionDialog = defaultKeyMaps.Chat.ShowCompletionDialog
		}
	}

	// Merge Global keymaps
	if userKeyMaps.Global == nil {
		userKeyMaps.Global = defaultKeyMaps.Global
	} else {
		if userKeyMaps.Global.ViewLogs == nil {
			userKeyMaps.Global.ViewLogs = defaultKeyMaps.Global.ViewLogs
		}
		if userKeyMaps.Global.Quit == nil {
			userKeyMaps.Global.Quit = defaultKeyMaps.Global.Quit
		}
		if userKeyMaps.Global.Help == nil {
			userKeyMaps.Global.Help = defaultKeyMaps.Global.Help
		}
		if userKeyMaps.Global.SwitchSession == nil {
			userKeyMaps.Global.SwitchSession = defaultKeyMaps.Global.SwitchSession
		}
		if userKeyMaps.Global.Commands == nil {
			userKeyMaps.Global.Commands = defaultKeyMaps.Global.Commands
		}
		if userKeyMaps.Global.FilePicker == nil {
			userKeyMaps.Global.FilePicker = defaultKeyMaps.Global.FilePicker
		}
		if userKeyMaps.Global.Models == nil {
			userKeyMaps.Global.Models = defaultKeyMaps.Global.Models
		}
		if userKeyMaps.Global.Theme == nil {
			userKeyMaps.Global.Theme = defaultKeyMaps.Global.Theme
		}
		if userKeyMaps.Global.Tools == nil {
			userKeyMaps.Global.Tools = defaultKeyMaps.Global.Tools
		}
	}

	// Merge Editor keymaps
	if userKeyMaps.Editor == nil {
		userKeyMaps.Editor = defaultKeyMaps.Editor
	} else {
		if userKeyMaps.Editor.Submit == nil {
			userKeyMaps.Editor.Submit = defaultKeyMaps.Editor.Submit
		}
		if userKeyMaps.Editor.Clear == nil {
			userKeyMaps.Editor.Clear = defaultKeyMaps.Editor.Clear
		}
	}

	// Merge Messages keymaps
	if userKeyMaps.Messages == nil {
		userKeyMaps.Messages = defaultKeyMaps.Messages
	} else {
		if userKeyMaps.Messages.HalfPageUp == nil {
			userKeyMaps.Messages.HalfPageUp = defaultKeyMaps.Messages.HalfPageUp
		}
		if userKeyMaps.Messages.HalfPageDown == nil {
			userKeyMaps.Messages.HalfPageDown = defaultKeyMaps.Messages.HalfPageDown
		}
	}

	// Merge Logs keymaps
	if userKeyMaps.Logs == nil {
		userKeyMaps.Logs = defaultKeyMaps.Logs
	} else {
		if userKeyMaps.Logs.Back == nil {
			userKeyMaps.Logs.Back = defaultKeyMaps.Logs.Back
		}
	}
}
