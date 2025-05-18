package completions

import (
	"context"

	"github.com/sst/opencode/internal/llm/tools"
	"github.com/sst/opencode/internal/lsp"
	"github.com/sst/opencode/internal/tui/components/dialog"
)

type lspCompletionProvider struct {
	prefix     string
	lspClients map[string]*lsp.Client
}

func (cg *lspCompletionProvider) GetId() string {
	return cg.prefix
}

func (cg *lspCompletionProvider) GetEntry() dialog.CompletionItemI {
	return dialog.NewCompletionItem(dialog.CompletionItem{
		Title: "Lsp Symbols",
		Value: "lsp",
	})
}

func (cg *lspCompletionProvider) GetChildEntries(query string) ([]dialog.CompletionItemI, error) {
	items := make([]dialog.CompletionItemI, 0, 1)

	symbols := tools.GetWorkspaceSymbols(context.Background(), query, cg.lspClients)

	for _, symbol := range symbols {
		item := dialog.NewCompletionItem(dialog.CompletionItem{
			Title: "Test symbol",
			Value: symbol,
		})

		items = append(items, item)
	}

	return items, nil
}

func NewLspCompleitonProvider(lspClients map[string]*lsp.Client) dialog.CompletionProvider {
	return &lspCompletionProvider{
		prefix:     "lsp",
		lspClients: lspClients,
	}
}
