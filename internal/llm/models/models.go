package models

import "maps"

type (
	ModelID       string
	ModelProvider string
)

type Model struct {
	ID                  ModelID       `json:"id"`
	Name                string        `json:"name"`
	Provider            ModelProvider `json:"provider"`
	APIModel            string        `json:"api_model"`
	CostPer1MIn         float64       `json:"cost_per_1m_in"`
	CostPer1MOut        float64       `json:"cost_per_1m_out"`
	CostPer1MInCached   float64       `json:"cost_per_1m_in_cached"`
	CostPer1MOutCached  float64       `json:"cost_per_1m_out_cached"`
	ContextWindow       int64         `json:"context_window"`
	DefaultMaxTokens    int64         `json:"default_max_tokens"`
	CanReason           bool          `json:"can_reason"`
	SupportsAttachments bool          `json:"supports_attachments"`
}

const (
	// ForTests
	ProviderMock ModelProvider = "__mock"
)

// Providers in order of popularity
var ProviderPopularity = map[ModelProvider]int{
	ProviderAnthropic:  1,
	ProviderOpenAI:     2,
	ProviderGemini:     3,
	ProviderGROQ:       4,
	ProviderOpenRouter: 5,
	ProviderBedrock:    6,
	ProviderAzure:      7,
	ProviderVertexAI:   8,
}

var SupportedModels = map[ModelID]Model{}

func init() {
	maps.Copy(SupportedModels, AnthropicModels)
	maps.Copy(SupportedModels, BedrockModels)
	maps.Copy(SupportedModels, OpenAIModels)
	maps.Copy(SupportedModels, GeminiModels)
	maps.Copy(SupportedModels, GroqModels)
	maps.Copy(SupportedModels, AzureModels)
	maps.Copy(SupportedModels, OpenRouterModels)
	maps.Copy(SupportedModels, XAIModels)
	maps.Copy(SupportedModels, VertexAIGeminiModels)
}

// AvailableProviders returns a list of all available providers
func AvailableProviders() ([]ModelProvider, map[ModelProvider]string) {
	providerLabels := make(map[ModelProvider]string)
	providerLabels[ProviderAnthropic] = "Anthropic"
	providerLabels[ProviderAzure] = "Azure"
	providerLabels[ProviderBedrock] = "Bedrock"
	providerLabels[ProviderGemini] = "Gemini"
	providerLabels[ProviderGROQ] = "Groq"
	providerLabels[ProviderOpenAI] = "OpenAI"
	providerLabels[ProviderOpenRouter] = "OpenRouter"
	providerLabels[ProviderXAI] = "xAI"

	providerList := make([]ModelProvider, 0, len(providerLabels))
	providerList = append(providerList, ProviderAnthropic)
	providerList = append(providerList, ProviderAzure)
	providerList = append(providerList, ProviderBedrock)
	providerList = append(providerList, ProviderGemini)
	providerList = append(providerList, ProviderGROQ)
	providerList = append(providerList, ProviderOpenAI)
	providerList = append(providerList, ProviderOpenRouter)
	providerList = append(providerList, ProviderXAI)

	return providerList, providerLabels
}

// AvailableModelsByProvider returns a list of all available models by provider
func AvailableModelsByProvider(provider ModelProvider) []Model {
	var modelMap map[ModelID]Model

	switch provider {
	default:
		modelMap = map[ModelID]Model{}
	case ProviderAnthropic:
		modelMap = AnthropicModels
	case ProviderAzure:
		modelMap = AzureModels
	case ProviderBedrock:
		modelMap = BedrockModels
	case ProviderGemini:
		modelMap = GeminiModels
	case ProviderGROQ:
		modelMap = GroqModels
	case ProviderOpenAI:
		modelMap = OpenAIModels
	case ProviderOpenRouter:
		modelMap = OpenRouterModels
	case ProviderXAI:
		modelMap = XAIModels
	}

	models := make([]Model, 0, len(modelMap))
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
