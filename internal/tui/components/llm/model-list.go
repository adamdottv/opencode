package llm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sst/opencode/internal/llm/models"
	utilComponents "github.com/sst/opencode/internal/tui/components/util"
	"github.com/sst/opencode/internal/tui/styles"
	"github.com/sst/opencode/internal/tui/theme"
)

type ModelListItem struct {
	Model models.Model
}

func (p ModelListItem) Render(selected bool, width int) string {
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

	title := itemStyle.Padding(0, 1).Render(p.Model.Name)
	return title
}

type ModelList struct {
	list   utilComponents.SimpleList[ModelListItem]
	models map[models.ModelProvider][]models.Model
}

func (p *ModelList) View() string {
	return p.list.View()
}

func (p *ModelList) Update(msg tea.Msg) (ModelList, tea.Cmd) {
	l, cmd := p.list.Update(msg)
	p.list = l.(utilComponents.SimpleList[ModelListItem])

	return *p, cmd
}

func BuildListItemsForProvider(provider models.ModelProvider) []ModelListItem {
	modelsByProvider := models.AvailableModelsByProvider()

	modelListItems := make([]ModelListItem, 0, len(modelsByProvider[provider]))
	for _, model := range modelsByProvider[provider] {
		modelListItems = append(modelListItems, ModelListItem{Model: model})
	}

	return modelListItems
}

func (p *ModelList) SetProvider(provider models.ModelProvider) {
	modelListItems := BuildListItemsForProvider(provider)

	p.list.SetItems(modelListItems)
}

func (p *ModelList) GetSelectedModel() ModelListItem {
	item, _ := p.list.GetSelectedItem()

	return item
}

type NewModelListOptions struct {
	AlphaNumericKeys *bool
	FallbackMsg      *string
	InitialProvider  models.ModelProvider
	MaxVisibleItems  *int
	Width            *int
}

func NewModelList(options NewModelListOptions) ModelList {
	var maxVisibleItems = 10
	if options.MaxVisibleItems != nil {
		maxVisibleItems = *options.MaxVisibleItems
	}

	var fallbackMsg = "No models found"
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

	listItems := BuildListItemsForProvider(options.InitialProvider)
	list := utilComponents.NewSimpleList(listItems, maxVisibleItems, fallbackMsg, useAlphaNumericKeys)
	list.SetMaxWidth(width)

	return ModelList{
		list: list,
	}
}
