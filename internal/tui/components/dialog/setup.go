package dialog

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sst/opencode/internal/llm/models"
	"github.com/sst/opencode/internal/tui/layout"
	"github.com/sst/opencode/internal/tui/styles"
	"github.com/sst/opencode/internal/tui/theme"
	"github.com/sst/opencode/internal/tui/util"
	"strings"
)

type SetupDialog interface {
	tea.Model
	layout.Bindings
}

type SetupStep string

const (
	Start          SetupStep = "start"
	SelectProvider SetupStep = "select-provider"
	SelectModel    SetupStep = "select-model"
	InputApiKey    SetupStep = "input-api-key"
)

type setupDialogCmp struct {
	currentModel        string
	currentProvider     string
	keys                setupMapping
	models              []models.Model
	providers           []models.ModelProvider
	providerLabels      map[models.ModelProvider]string
	selectedModelIdx    int
	selectedProviderIdx int
	step                SetupStep
	textInput           textinput.Model
	textInputError      string
	width               int
}

type setupMapping struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Escape key.Binding
}

var setupKeys = setupMapping{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "prev"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "next"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "next"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

func (s *setupDialogCmp) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (s *setupDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cursor.BlinkMsg:
		if s.step == InputApiKey {
			// textinput.Update() does not work to make the cursor blink
			// we need to manually toggle the blink state
			s.textInput.Cursor.Blink = !s.textInput.Cursor.Blink
			return s, nil
		}
	case tea.KeyMsg:
		if s.step == Start && key.Matches(msg, setupKeys.Enter) {
			s.step = SelectProvider
			return s, nil
		}

		if s.step == SelectProvider {
			switch {
			case key.Matches(msg, setupKeys.Up):
				s.selectedProviderIdx--
				if s.selectedProviderIdx < 0 {
					s.selectedProviderIdx = len(s.providers) - 1
				}
			case key.Matches(msg, setupKeys.Down):
				s.selectedProviderIdx++
				if s.selectedProviderIdx >= len(s.providers) {
					s.selectedProviderIdx = 0
				}
			case key.Matches(msg, setupKeys.Enter):
				s.models = models.AvailableModelsByProvider(s.providers[s.selectedProviderIdx])
				s.step = SelectModel
			case key.Matches(msg, setupKeys.Escape):
				s.step = Start
			}

			return s, nil
		}

		if s.step == SelectModel {
			switch {
			case key.Matches(msg, setupKeys.Up):
				s.selectedModelIdx--
				if s.selectedModelIdx < 0 {
					s.selectedModelIdx = len(s.providers) - 1
				}
			case key.Matches(msg, setupKeys.Down):
				s.selectedModelIdx++
				if s.selectedModelIdx >= len(s.providers) {
					s.selectedProviderIdx = 0
				}
			case key.Matches(msg, setupKeys.Enter):
				s.step = InputApiKey
				s.textInput.Focus()
			case key.Matches(msg, setupKeys.Escape):
				s.selectedModelIdx = 0
				s.step = SelectProvider
			}

			return s, nil
		}

		if s.step == InputApiKey {
			switch {
			case key.Matches(msg, setupKeys.Escape):
				s.step = SelectModel

			case key.Matches(msg, setupKeys.Enter):
				if s.textInput.Value() == "" {
					s.textInputError = "Field cannot be empty"
					return s, nil
				}

				return s, util.CmdHandler(CloseSetupDialogMsg{
					Provider: s.providers[s.selectedProviderIdx],
					Model:    s.models[s.selectedModelIdx],
					APIKey:   s.textInput.Value(),
				})
			}

			var cmd tea.Cmd
			var cmds []tea.Cmd

			s.textInput, cmd = s.textInput.Update(msg)
			cmds = append(cmds, cmd)

			return s, tea.Batch(cmds...)
		}
	}

	return s, nil
}

func (s *setupDialogCmp) View() string {
	switch s.step {
	default:
		return s.RenderSetupStep()
	case Start:
		return s.RenderSetupStep()
	case SelectProvider:
		return s.RenderSelectProviderStep()
	case SelectModel:
		return s.RenderSelectModelStep()
	case InputApiKey:
		return s.RenderInputApiKeyStep()
	}
}

func (s *setupDialogCmp) renderAndPadLine(text string, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	spacerStyle := baseStyle.Background(t.Background())
	return text + spacerStyle.Render(strings.Repeat(" ", width-lipgloss.Width(text)))
}

func (s *setupDialogCmp) renderHelp() string {
	// We have to render the help manually due to artifacting when using help.View(s.keys)
	// this is a bug with the bubbletea/help package
	t := theme.CurrentTheme()
	sepStyle := styles.BaseStyle().Foreground(t.Primary())
	sep := sepStyle.Render(" • ")

	keyStyle := styles.BaseStyle().Foreground(t.Text())
	descStyle := styles.BaseStyle().Foreground(t.TextMuted())
	space := styles.BaseStyle().Foreground(t.Background()).Render(" ")
	key1 := keyStyle.Render(s.keys.Escape.Help().Key)
	desc1 := descStyle.Render(s.keys.Escape.Help().Desc)
	key2 := keyStyle.Render(s.keys.Enter.Help().Key)
	desc2 := descStyle.Render(s.keys.Enter.Help().Desc)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		key1,
		space,
		desc1,
		sep,
		key2,
		space,
		desc2,
	)
}

func (s *setupDialogCmp) RenderSetupStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	nextStyle := baseStyle
	nextStyle = nextStyle.Background(t.Primary()).Foreground(t.Background())
	spacerStyle := baseStyle.Background(t.Background())

	nextButton := nextStyle.Padding(0, 1).Render("Proceed")

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, nextButton)

	line1 := "✨ Welcome to OpenCode"
	line2 := "Your AI-powered coding companion is almost ready!"
	line3 := "Let's get you set up with your preferred AI provider, model, and API key."

	width := lipgloss.Width(line3)
	remainingWidth := width - lipgloss.Width(buttons)
	if remainingWidth > 0 {
		buttons = spacerStyle.Render(strings.Repeat(" ", remainingWidth)) + buttons
	}

	title := baseStyle.
		Background(t.Background()).
		Foreground(t.Primary()).
		Bold(true).
		Render("Setup Wizard")

	content := baseStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			s.renderAndPadLine(title, width),
			"",
			s.renderAndPadLine(line1, width),
			"",
			s.renderAndPadLine(line2, width),
			"",
			s.renderAndPadLine(line3, width),
			"",
			buttons,
		),
	)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

func (s *setupDialogCmp) RenderSelectProviderStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate max width needed for provider names
	maxWidth := 36
	for _, providerName := range s.providers {
		if len(providerName) > maxWidth {
			maxWidth = len(providerName)
		}
	}

	helpText := s.renderHelp()
	helpWidth := lipgloss.Width(helpText)
	maxWidth = max(maxWidth, helpWidth)

	// Add padding to help
	remainingWidth := maxWidth - lipgloss.Width(helpText)
	if remainingWidth > 0 {
		helpText = strings.Repeat(" ", remainingWidth) + helpText
	}

	// Build the provider list
	providerItems := make([]string, 0, len(s.providers))
	for i, provider := range s.providers {
		itemStyle := baseStyle.Width(maxWidth)

		if i == s.selectedProviderIdx {
			itemStyle = itemStyle.
				Background(t.Primary()).
				Foreground(t.Background()).
				Bold(true)
		}

		providerItems = append(providerItems, itemStyle.Padding(0, 1).Render(s.providerLabels[provider]))
	}

	title := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Width(maxWidth).
		Padding(0, 1).
		Render("Select Provider")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		baseStyle.Width(maxWidth).Render(""),
		baseStyle.Width(maxWidth).Render(lipgloss.JoinVertical(lipgloss.Left, providerItems...)),
		baseStyle.Width(maxWidth).Render("\n\n"),
		baseStyle.Width(maxWidth).Render(helpText),
	)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

func (s *setupDialogCmp) RenderSelectModelStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate max width needed for model names
	maxWidth := 36
	for _, model := range s.models {
		if len(model.Name) > maxWidth {
			maxWidth = len(model.Name)
		}
	}

	helpText := s.renderHelp()
	helpWidth := lipgloss.Width(helpText)
	maxWidth = max(maxWidth, helpWidth)

	// Add padding to help
	remainingWidth := maxWidth - lipgloss.Width(helpText)
	if remainingWidth > 0 {
		helpText = strings.Repeat(" ", remainingWidth) + helpText
	}

	// Build the model list
	modelItems := make([]string, 0, len(s.models))
	for i, model := range s.models {
		itemStyle := baseStyle.Width(maxWidth)

		if i == s.selectedModelIdx {
			itemStyle = itemStyle.
				Background(t.Primary()).
				Foreground(t.Background()).
				Bold(true)
		}

		modelItems = append(modelItems, itemStyle.Padding(0, 1).Render(model.Name))
	}

	title := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Width(maxWidth).
		Padding(0, 1).
		Render("Select Model")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		baseStyle.Width(maxWidth).Render(""),
		baseStyle.Width(maxWidth).Render(lipgloss.JoinVertical(lipgloss.Left, modelItems...)),
		baseStyle.Width(maxWidth).Render("\n\n"),
		baseStyle.Width(maxWidth).Render(helpText),
	)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

func (s *setupDialogCmp) RenderInputApiKeyStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate width needed for content
	maxWidth := 60 // Width for explanation text

	helpText := s.renderHelp()
	helpWidth := lipgloss.Width(helpText)
	maxWidth = max(60, helpWidth) // Limit width to avoid overflow

	// Add padding to help
	remainingWidth := maxWidth - lipgloss.Width(helpText)
	if remainingWidth > 0 {
		helpText = strings.Repeat(" ", remainingWidth) + helpText
	}

	title := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Width(maxWidth).
		Padding(0, 1).
		Render("API Key")

	inputField := baseStyle.
		Foreground(t.Text()).
		Width(maxWidth).
		Padding(1, 1).
		Render(s.textInput.View())

	errorStyle := baseStyle.Foreground(t.Error()).PaddingLeft(1)
	errorText := ""
	if s.textInputError != "" {
		errorText = errorStyle.Render(s.renderAndPadLine(s.textInputError, maxWidth-1))
	}

	maxWidth = min(maxWidth, s.width-10)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		inputField,
		errorText,
		baseStyle.Width(maxWidth).Render(helpText),
	)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Background(t.Background()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

func (s *setupDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(setupKeys)
}

func NewSetupDialogCmp() SetupDialog {
	t := theme.CurrentTheme()

	ti := textinput.New()
	ti.Placeholder = "Enter API Key..."
	ti.Width = 56
	ti.Prompt = ""
	ti.PlaceholderStyle = ti.PlaceholderStyle.Background(t.Background())
	ti.PromptStyle = ti.PromptStyle.Background(t.Background())
	ti.TextStyle = ti.TextStyle.Background(t.Background())

	providers, providerLabels := models.AvailableProviders()

	return &setupDialogCmp{
		keys:           setupKeys,
		providers:      providers,
		providerLabels: providerLabels,
		step:           Start,
		textInput:      ti,
	}
}

// CloseSetupDialogMsg is a message that is sent when the init dialog is closed.
type CloseSetupDialogMsg struct {
	APIKey   string
	Model    models.Model
	Provider models.ModelProvider
}

// ShowSetupDialogMsg is a message that is sent to show the init dialog.
type ShowSetupDialogMsg struct {
	Show bool
}
