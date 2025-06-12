package util

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	StatusMsg struct {
		Text string
		TTL  time.Duration
	}
	ClearStatusMsg struct{}
	FocusMsg       struct {
		Focused bool
	}
)

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func Clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
