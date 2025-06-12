package util

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseFocusMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    tea.KeyMsg
		expected bool
		focused  bool
	}{
		{
			name:     "focus in",
			input:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\x1b[I")},
			expected: true,
			focused:  true,
		},
		{
			name:     "focus out",
			input:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("\x1b[O")},
			expected: true,
			focused:  false,
		},
		{
			name:     "regular key",
			input:    tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			expected: false,
			focused:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, msg := ParseFocusMessage(tt.input)
			if ok != tt.expected {
				t.Errorf("ParseFocusMessage() ok = %v, want %v", ok, tt.expected)
			}
			if ok && msg.Focused != tt.focused {
				t.Errorf("ParseFocusMessage() focused = %v, want %v", msg.Focused, tt.focused)
			}
		})
	}
}

func TestFocusTracker(t *testing.T) {
	tracker := NewFocusTracker(nil)

	// Test initial state
	if !tracker.IsFocused() {
		t.Error("FocusTracker should default to focused")
	}

	// Test focus change
	tracker.HandleFocusEvent(false)
	if tracker.IsFocused() {
		t.Error("FocusTracker should be unfocused after HandleFocusEvent(false)")
	}

	tracker.HandleFocusEvent(true)
	if !tracker.IsFocused() {
		t.Error("FocusTracker should be focused after HandleFocusEvent(true)")
	}
}
