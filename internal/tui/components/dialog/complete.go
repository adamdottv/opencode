package dialog

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sst/opencode/internal/lsp"
	"github.com/sst/opencode/internal/lsp/protocol"
	"github.com/sst/opencode/internal/status"
	utilComponents "github.com/sst/opencode/internal/tui/components/util"
	"github.com/sst/opencode/internal/tui/layout"
	"github.com/sst/opencode/internal/tui/styles"
	"github.com/sst/opencode/internal/tui/theme"
	"github.com/sst/opencode/internal/tui/util"
)

type CompletionItem struct {
	title string
	Title string
	Value string
}

type CompletionItemI interface {
	utilComponents.SimpleListItem
	GetValue() string
	DisplayValue() string
}

func (ci *CompletionItem) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	itemStyle := baseStyle.
		Width(width).
		Padding(0, 1)

	if selected {
		itemStyle = itemStyle.
			Background(t.Background()).
			Foreground(t.Primary()).
			Bold(true)
	}

	title := itemStyle.Render(
		ci.GetValue(),
	)

	return title
}

func (ci *CompletionItem) DisplayValue() string {
	return ci.Title
}

func (ci *CompletionItem) GetValue() string {
	return ci.Value
}

func NewCompletionItem(completionItem CompletionItem) CompletionItemI {
	return &completionItem
}

type CompletionProvider interface {
	GetId() string
	GetEntry() CompletionItemI
	GetChildEntries(query string) ([]CompletionItemI, error)
}

type CompletionSelectedMsg struct {
	SearchString    string
	CompletionValue string
}

type CompletionDialogCompleteItemMsg struct {
	Value string
}

type CompletionDialogCloseMsg struct{}

type CompletionDialog interface {
	tea.Model
	layout.Bindings
	SetWidth(width int)
}

type completionDialogCmp struct {
	query                      string
	completionProviders        []CompletionProvider
	selectedCompletionProvider int
	width                      int
	height                     int
	pseudoSearchTextArea       textarea.Model
	listView                   utilComponents.SimpleList[CompletionItemI]
}

type completionDialogKeyMap struct {
	Complete key.Binding
	Cancel   key.Binding
}

var completionDialogKeys = completionDialogKeyMap{
	Complete: key.NewBinding(
		key.WithKeys("tab", "enter"),
	),
	Cancel: key.NewBinding(
		key.WithKeys(" ", "esc", "backspace"),
	),
}

func (c *completionDialogCmp) Init() tea.Cmd {
	return nil
}

func (c *completionDialogCmp) complete(item CompletionItemI) tea.Cmd {
	value := c.pseudoSearchTextArea.Value()

	if value == "" {
		return nil
	}

	return tea.Batch(
		util.CmdHandler(CompletionSelectedMsg{
			SearchString:    value,
			CompletionValue: item.GetValue(),
		}),
		c.close(),
	)
}

func (c *completionDialogCmp) close() tea.Cmd {
	c.pseudoSearchTextArea.Reset()
	c.pseudoSearchTextArea.Blur()
	c.selectedCompletionProvider = -1

	items := make([]CompletionItemI, 0)

	for _, provider := range c.completionProviders {
		items = append(items, provider.GetEntry())
	}

	c.listView.SetItems(items)

	return util.CmdHandler(CompletionDialogCloseMsg{})
}

func (c *completionDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if c.pseudoSearchTextArea.Focused() {

			if !key.Matches(msg, completionDialogKeys.Complete) {

				var cmd tea.Cmd
				c.pseudoSearchTextArea, cmd = c.pseudoSearchTextArea.Update(msg)
				cmds = append(cmds, cmd)

				var query string
				query = c.pseudoSearchTextArea.Value()
				if query != "" {
					query = query[1:]

					if query != c.query {
						completionsItems := make([]CompletionItemI, 0)
						if query != "" {
							for _, provider := range c.completionProviders {
								items, err := provider.GetChildEntries(query)
								if err != nil {
									status.Error(err.Error())
								}
								completionsItems = append(completionsItems, items...)
							}
						} else {
							if c.selectedCompletionProvider == -1 {
								for _, provider := range c.completionProviders {
									items := provider.GetEntry()
									completionsItems = append(completionsItems, items)
								}
							} else {
								provider := c.completionProviders[c.selectedCompletionProvider]
								items, err := provider.GetChildEntries(query)
								if err != nil {
									status.Error(err.Error())
								}

								completionsItems = append(completionsItems, items...)
							}
						}
						c.listView.SetItems(completionsItems)
						c.query = query
					}

				}

				u, cmd := c.listView.Update(msg)
				c.listView = u.(utilComponents.SimpleList[CompletionItemI])

				cmds = append(cmds, cmd)
			}

			switch {
			case key.Matches(msg, completionDialogKeys.Complete):
				item, i := c.listView.GetSelectedItem()
				if i == -1 {
					return c, nil
				}

				var cmd tea.Cmd = nil

				if c.selectedCompletionProvider == -1 {
					c.selectedCompletionProvider = i

					provider := c.completionProviders[i]

					items, err := provider.GetChildEntries("")
					if err != nil {
						status.Error(err.Error())
					}

					c.listView.SetItems(items)
				} else {
					cmd = c.complete(item)
				}

				return c, cmd
			case key.Matches(msg, completionDialogKeys.Cancel):
				// Only close on backspace when there are no characters left
				if msg.String() != "backspace" || len(c.pseudoSearchTextArea.Value()) <= 0 {
					return c, c.close()
				}
			}

			return c, tea.Batch(cmds...)
		} else {
			c.pseudoSearchTextArea.SetValue(msg.String())
			return c, c.pseudoSearchTextArea.Focus()
		}
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	}

	return c, tea.Batch(cmds...)
}

func (c *completionDialogCmp) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	maxWidth := 40

	completions := c.listView.GetItems()

	for _, cmd := range completions {
		title := cmd.DisplayValue()
		if len(title) > maxWidth-4 {
			maxWidth = len(title) + 4
		}
	}

	c.listView.SetMaxWidth(maxWidth)

	return baseStyle.Padding(0, 0).
		Border(lipgloss.NormalBorder()).
		BorderBottom(false).
		BorderRight(false).
		BorderLeft(false).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(c.width).
		Render(c.listView.View())
}

func (c *completionDialogCmp) SetWidth(width int) {
	c.width = width
}

func (c *completionDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(completionDialogKeys)
}

func NewCompletionDialogCmp(completionProvider []CompletionProvider) CompletionDialog {
	ti := textarea.New()

	items := make([]CompletionItemI, 0)

	for _, provider := range completionProvider {
		items = append(items, provider.GetEntry())
	}

	li := utilComponents.NewSimpleList(
		items,
		7,
		"No file matches found",
		false,
	)

	return &completionDialogCmp{
		query:                      "",
		completionProviders:        completionProvider,
		selectedCompletionProvider: -1,
		pseudoSearchTextArea:       ti,
		listView:                   li,
	}
}

func getDocumentSymbols(ctx context.Context, filePath string, lsps map[string]*lsp.Client) string {
	var results []string

	for lspName, client := range lsps {
		// Create document symbol params
		uri := fmt.Sprintf("file://%s", filePath)
		symbolParams := protocol.DocumentSymbolParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentUri(uri),
			},
		}

		// Get document symbols
		symbolResult, err := client.DocumentSymbol(ctx, symbolParams)
		if err != nil {
			results = append(results, fmt.Sprintf("Error from %s: %s", lspName, err))
			continue
		}

		// Process the symbol result
		symbols := processDocumentSymbolResult(symbolResult)
		if len(symbols) == 0 {
			results = append(results, fmt.Sprintf("No symbols found by %s", lspName))
			continue
		}

		// Format the symbols
		results = append(results, fmt.Sprintf("Symbols found by %s:", lspName))
		for _, symbol := range symbols {
			results = append(results, formatSymbol(symbol, 1))
		}
	}

	if len(results) == 0 {
		return "No symbols found in the specified file."
	}

	return strings.Join(results, "\n")
}

func processDocumentSymbolResult(result protocol.Or_Result_textDocument_documentSymbol) []SymbolInfo {
	var symbols []SymbolInfo

	switch v := result.Value.(type) {
	case []protocol.SymbolInformation:
		for _, si := range v {
			symbols = append(symbols, SymbolInfo{
				Name:     si.Name,
				Kind:     symbolKindToString(si.Kind),
				Location: locationToString(si.Location),
				Children: nil,
			})
		}
	case []protocol.DocumentSymbol:
		for _, ds := range v {
			symbols = append(symbols, documentSymbolToSymbolInfo(ds))
		}
	}

	return symbols
}

// SymbolInfo represents a symbol in a document
type SymbolInfo struct {
	Name     string
	Kind     string
	Location string
	Children []SymbolInfo
}

func documentSymbolToSymbolInfo(symbol protocol.DocumentSymbol) SymbolInfo {
	info := SymbolInfo{
		Name: symbol.Name,
		Kind: symbolKindToString(symbol.Kind),
		Location: fmt.Sprintf("Line %d-%d",
			symbol.Range.Start.Line+1,
			symbol.Range.End.Line+1),
		Children: []SymbolInfo{},
	}

	for _, child := range symbol.Children {
		info.Children = append(info.Children, documentSymbolToSymbolInfo(child))
	}

	return info
}

func locationToString(location protocol.Location) string {
	return fmt.Sprintf("Line %d-%d",
		location.Range.Start.Line+1,
		location.Range.End.Line+1)
}

func symbolKindToString(kind protocol.SymbolKind) string {
	if kindStr, ok := protocol.TableKindMap[kind]; ok {
		return kindStr
	}
	return "Unknown"
}

func formatSymbol(symbol SymbolInfo, level int) string {
	indent := strings.Repeat("  ", level)
	result := fmt.Sprintf("%s- %s (%s) %s", indent, symbol.Name, symbol.Kind, symbol.Location)

	var childResults []string
	for _, child := range symbol.Children {
		childResults = append(childResults, formatSymbol(child, level+1))
	}

	if len(childResults) > 0 {
		return result + "\n" + strings.Join(childResults, "\n")
	}

	return result
}
