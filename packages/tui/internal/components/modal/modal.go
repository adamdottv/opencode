package modal

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

// CloseModalMsg is a message to signal that the active modal should be closed.
type CloseModalMsg struct{}

// Modal is a reusable modal component that handles frame rendering and overlay placement
type Modal struct {
	width      int
	height     int
	title      string
	maxWidth   int
	maxHeight  int
	fitContent bool
}

// ModalOption is a function that configures a Modal
type ModalOption func(*Modal)

// WithTitle sets the modal title
func WithTitle(title string) ModalOption {
	return func(m *Modal) {
		m.title = title
	}
}

// WithMaxWidth sets the maximum width
func WithMaxWidth(width int) ModalOption {
	return func(m *Modal) {
		m.maxWidth = width
		m.fitContent = false
	}
}

// WithMaxHeight sets the maximum height
func WithMaxHeight(height int) ModalOption {
	return func(m *Modal) {
		m.maxHeight = height
	}
}

func WithFitContent(fit bool) ModalOption {
	return func(m *Modal) {
		m.fitContent = fit
	}
}

// New creates a new Modal with the given options
func New(opts ...ModalOption) *Modal {
	m := &Modal{
		maxWidth:   0,
		maxHeight:  0,
		fitContent: true,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Render renders the modal centered on the screen
func (m *Modal) Render(contentView string, background string) string {
	t := theme.CurrentTheme()

	outerWidth := layout.Current.Container.Width - 8
	if m.maxWidth > 0 && outerWidth > m.maxWidth {
		outerWidth = m.maxWidth
	}

	if m.fitContent {
		titleWidth := lipgloss.Width(m.title)
		contentWidth := lipgloss.Width(contentView)
		largestWidth := max(titleWidth+2, contentWidth)
		outerWidth = largestWidth + 6
	}

	innerWidth := outerWidth - 4

	// Base style for the modal
	baseStyle := styles.BaseStyle().
		Background(t.BackgroundElement()).
		Foreground(t.TextMuted())

	// Add title if provided
	var finalContent string
	if m.title != "" {
		titleStyle := baseStyle.
			Foreground(t.Primary()).
			Bold(true).
			Width(innerWidth).
			Padding(0, 1)

		titleView := titleStyle.Render(m.title)
		finalContent = lipgloss.JoinVertical(
			lipgloss.Left,
			titleView,
			contentView,
		)
	} else {
		finalContent = contentView
	}

	modalStyle := baseStyle.
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderRight(true).
		BorderLeftForeground(t.BackgroundSubtle()).
		BorderLeftBackground(t.Background()).
		BorderRightForeground(t.BackgroundSubtle()).
		BorderRightBackground(t.Background())

	modalView := modalStyle.
		Width(outerWidth).
		Render(finalContent)

	// Calculate position for centering
	bgHeight := lipgloss.Height(background)
	bgWidth := lipgloss.Width(background)
	modalHeight := lipgloss.Height(modalView)
	modalWidth := lipgloss.Width(modalView)

	row := (bgHeight - modalHeight) / 2
	col := (bgWidth - modalWidth) / 2

	return layout.PlaceOverlay(
		col,
		row,
		modalView,
		background,
	)
}
