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

// AvailableProviders returns a list of all available providers
func AvailableProviders() ([]models.ModelProvider, map[models.ModelProvider]string) {
	providerLabels := make(map[models.ModelProvider]string)
	providerLabels[models.ProviderAnthropic] = "Anthropic"
	providerLabels[models.ProviderAzure] = "Azure"
	providerLabels[models.ProviderBedrock] = "Bedrock"
	providerLabels[models.ProviderGemini] = "Gemini"
	providerLabels[models.ProviderGROQ] = "Groq"
	providerLabels[models.ProviderOpenAI] = "OpenAI"
	providerLabels[models.ProviderOpenRouter] = "OpenRouter"
	providerLabels[models.ProviderXAI] = "xAI"

	providerList := make([]models.ModelProvider, 0, len(providerLabels))
	providerList = append(providerList, models.ProviderAnthropic)
	providerList = append(providerList, models.ProviderAzure)
	providerList = append(providerList, models.ProviderBedrock)
	providerList = append(providerList, models.ProviderGemini)
	providerList = append(providerList, models.ProviderGROQ)
	providerList = append(providerList, models.ProviderOpenAI)
	providerList = append(providerList, models.ProviderOpenRouter)
	providerList = append(providerList, models.ProviderXAI)

	return providerList, providerLabels
}

// AvailableModelsByProvider returns a list of all available models by provider
func AvailableModelsByProvider(provider models.ModelProvider) []models.Model {
	var modelMap map[models.ModelID]models.Model

	switch provider {
	default:
		modelMap = map[models.ModelID]models.Model{}
	case models.ProviderAnthropic:
		modelMap = models.AnthropicModels
	case models.ProviderAzure:
		modelMap = models.AzureModels
	case models.ProviderBedrock:
		modelMap = models.BedrockModels
	case models.ProviderGemini:
		modelMap = models.GeminiModels
	case models.ProviderGROQ:
		modelMap = models.GroqModels
	case models.ProviderOpenAI:
		modelMap = models.OpenAIModels
	case models.ProviderOpenRouter:
		modelMap = models.OpenRouterModels
	case models.ProviderXAI:
		modelMap = models.XAIModels
	}

	models := make([]models.Model, 0, len(modelMap))
	for _, model := range modelMap {
		models = append(models, model)
	}

	// Sort models by alphabetical order
	for i := 0; i < len(models)-1; i++ {
		for j := i + 1; j < len(models); j++ {
			if models[i].Name > models[j].Name {
				models[i], models[j] = models[j], models[i]
			}
		}
	}

	return models
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

func (q *setupDialogCmp) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (q *setupDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cursor.BlinkMsg:
		if q.step == InputApiKey {
			// textinput.Update() does not work to make the cursor blink
			// we need to manually toggle the blink state
			q.textInput.Cursor.Blink = !q.textInput.Cursor.Blink
			return q, nil
		}
	case tea.KeyMsg:
		if q.step == Start && key.Matches(msg, setupKeys.Enter) {
			q.step = SelectProvider
			return q, nil
		}

		if q.step == SelectProvider {
			switch {
			case key.Matches(msg, setupKeys.Up):
				q.selectedProviderIdx--
				if q.selectedProviderIdx < 0 {
					q.selectedProviderIdx = len(q.providers) - 1
				}
			case key.Matches(msg, setupKeys.Down):
				q.selectedProviderIdx++
				if q.selectedProviderIdx >= len(q.providers) {
					q.selectedProviderIdx = 0
				}
			case key.Matches(msg, setupKeys.Enter):
				q.models = AvailableModelsByProvider(q.providers[q.selectedProviderIdx])
				q.step = SelectModel
			case key.Matches(msg, setupKeys.Escape):
				q.step = Start
			}

			return q, nil
		}

		if q.step == SelectModel {
			switch {
			case key.Matches(msg, setupKeys.Up):
				q.selectedModelIdx--
				if q.selectedModelIdx < 0 {
					q.selectedModelIdx = len(q.providers) - 1
				}
			case key.Matches(msg, setupKeys.Down):
				q.selectedModelIdx++
				if q.selectedModelIdx >= len(q.providers) {
					q.selectedProviderIdx = 0
				}
			case key.Matches(msg, setupKeys.Enter):
				q.step = InputApiKey
				q.textInput.Focus()
			case key.Matches(msg, setupKeys.Escape):
				q.selectedModelIdx = 0
				q.step = SelectProvider
			}

			return q, nil
		}

		if q.step == InputApiKey {
			switch {
			case key.Matches(msg, setupKeys.Escape):
				q.step = SelectModel

			case key.Matches(msg, setupKeys.Enter):
				if q.textInput.Value() == "" {
					q.textInputError = "Field cannot be empty"
					return q, nil
				}

				return q, util.CmdHandler(CloseSetupDialogMsg{
					Provider: q.providers[q.selectedProviderIdx],
					Model:    q.models[q.selectedModelIdx],
					APIKey:   q.textInput.Value(),
				})
			}

			var cmd tea.Cmd
			var cmds []tea.Cmd

			q.textInput, cmd = q.textInput.Update(msg)
			cmds = append(cmds, cmd)

			return q, tea.Batch(cmds...)
		}
	}

	return q, nil
}

func (q *setupDialogCmp) View() string {
	switch q.step {
	default:
		return q.RenderSetupStep()
	case Start:
		return q.RenderSetupStep()
	case SelectProvider:
		return q.RenderSelectProviderStep()
	case SelectModel:
		return q.RenderSelectModelStep()
	case InputApiKey:
		return q.RenderInputApiKeyStep()
	}
}

func (q *setupDialogCmp) renderAndPadLine(text string, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	spacerStyle := baseStyle.Background(t.Background())
	return text + spacerStyle.Render(strings.Repeat(" ", width-lipgloss.Width(text)))
}

func (q *setupDialogCmp) renderHelp() string {
	// We have to render the help manually due to artifacting when using help.View(q.keys)
	// this is a bug with the bubbletea/help package
	t := theme.CurrentTheme()
	sepStyle := styles.BaseStyle().Foreground(t.Primary())
	sep := sepStyle.Render(" • ")

	keyStyle := styles.BaseStyle().Foreground(t.Text())
	descStyle := styles.BaseStyle().Foreground(t.TextMuted())
	space := styles.BaseStyle().Foreground(t.Background()).Render(" ")
	key1 := keyStyle.Render(q.keys.Escape.Help().Key)
	desc1 := descStyle.Render(q.keys.Escape.Help().Desc)
	key2 := keyStyle.Render(q.keys.Enter.Help().Key)
	desc2 := descStyle.Render(q.keys.Enter.Help().Desc)

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

func (q *setupDialogCmp) RenderSetupStep() string {
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
			q.renderAndPadLine(title, width),
			"",
			q.renderAndPadLine(line1, width),
			"",
			q.renderAndPadLine(line2, width),
			"",
			q.renderAndPadLine(line3, width),
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

func (q *setupDialogCmp) RenderSelectProviderStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate max width needed for provider names
	maxWidth := 36
	for _, providerName := range q.providers {
		if len(providerName) > maxWidth {
			maxWidth = len(providerName)
		}
	}

	helpText := q.renderHelp()
	helpWidth := lipgloss.Width(helpText)
	maxWidth = max(maxWidth, helpWidth)

	// Add padding to help
	remainingWidth := maxWidth - lipgloss.Width(helpText)
	if remainingWidth > 0 {
		helpText = strings.Repeat(" ", remainingWidth) + helpText
	}

	// Build the provider list
	providerItems := make([]string, 0, len(q.providers))
	for i, provider := range q.providers {
		itemStyle := baseStyle.Width(maxWidth)

		if i == q.selectedProviderIdx {
			itemStyle = itemStyle.
				Background(t.Primary()).
				Foreground(t.Background()).
				Bold(true)
		}

		providerItems = append(providerItems, itemStyle.Padding(0, 1).Render(q.providerLabels[provider]))
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

func (q *setupDialogCmp) RenderSelectModelStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate max width needed for model names
	maxWidth := 36
	for _, model := range q.models {
		if len(model.Name) > maxWidth {
			maxWidth = len(model.Name)
		}
	}

	helpText := q.renderHelp()
	helpWidth := lipgloss.Width(helpText)
	maxWidth = max(maxWidth, helpWidth)

	// Add padding to help
	remainingWidth := maxWidth - lipgloss.Width(helpText)
	if remainingWidth > 0 {
		helpText = strings.Repeat(" ", remainingWidth) + helpText
	}

	// Build the model list
	modelItems := make([]string, 0, len(q.models))
	for i, model := range q.models {
		itemStyle := baseStyle.Width(maxWidth)

		if i == q.selectedModelIdx {
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

func (q *setupDialogCmp) RenderInputApiKeyStep() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Calculate width needed for content
	maxWidth := 60 // Width for explanation text

	helpText := q.renderHelp()
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
		Render(q.textInput.View())

	errorStyle := baseStyle.Foreground(t.Error()).PaddingLeft(1)
	errorText := ""
	if q.textInputError != "" {
		errorText = errorStyle.Render(q.renderAndPadLine(q.textInputError, maxWidth-1))
	}

	maxWidth = min(maxWidth, q.width-10)

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

func (q *setupDialogCmp) BindingKeys() []key.Binding {
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

	providers, providerLabels := AvailableProviders()

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
