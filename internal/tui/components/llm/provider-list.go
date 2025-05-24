package llm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sst/opencode/internal/llm/models"
	utilComponents "github.com/sst/opencode/internal/tui/components/util"
	"github.com/sst/opencode/internal/tui/styles"
	"github.com/sst/opencode/internal/tui/theme"
)

type ProviderListItem struct {
	Label string
	Name  models.ModelProvider
}

func (p ProviderListItem) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	descStyle := baseStyle.Width(width).Foreground(t.TextMuted())
	itemStyle := baseStyle.Width(width).
		Background(t.Background()).
		Foreground(t.Text())

	if selected {
		itemStyle = itemStyle.
			Background(t.Primary()).
			Foreground(t.Background()).
			Bold(true)
		descStyle = descStyle.
			Background(t.Primary()).
			Foreground(t.Background())
	}

	title := itemStyle.Padding(0, 1).Render(p.Label)
	return title
}

type ProviderList struct {
	list utilComponents.SimpleList[ProviderListItem]
}

func (p *ProviderList) View() string {
	return p.list.View()
}

func (p *ProviderList) Update(msg tea.Msg) (ProviderList, tea.Cmd) {
	l, cmd := p.list.Update(msg)
	p.list = l.(utilComponents.SimpleList[ProviderListItem])

	return *p, cmd
}

func (p *ProviderList) GetSelectedProvider() ProviderListItem {
	item, _ := p.list.GetSelectedItem()

	return item
}

type NewProviderListOptions struct {
	AlphaNumericKeys *bool
	FallbackMsg      *string
	MaxVisibleItems  *int
	Width            *int
}

func NewProviderList(options NewProviderListOptions) ProviderList {
	providers, providerLabels := models.AvailableProviders()

	providerListItems := make([]ProviderListItem, 0, len(providers))
	for _, provider := range providers {
		providerListItems = append(providerListItems, ProviderListItem{Label: providerLabels[provider], Name: provider})
	}

	var maxVisibleItems = len(providers)
	if options.MaxVisibleItems != nil {
		maxVisibleItems = *options.MaxVisibleItems
	}

	var fallbackMsg = "No provider found"
	if options.FallbackMsg != nil {
		fallbackMsg = *options.FallbackMsg
	}

	var useAlphaNumericKeys = false
	if options.AlphaNumericKeys != nil {
		useAlphaNumericKeys = *options.AlphaNumericKeys
	}

	var width = 36
	if options.Width != nil {
		width = *options.Width
	}

	list := utilComponents.NewSimpleList(providerListItems, maxVisibleItems, fallbackMsg, useAlphaNumericKeys)
	list.SetMaxWidth(width)

	return ProviderList{
		list: list,
	}
}
