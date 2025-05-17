package setup

import (
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/llm/models"
	"log/slog"
)

// Global variable to track if setup is complete
var setupComplete = false

// IsSetupComplete checks if the setup is complete
func IsSetupComplete() bool {
	return setupComplete
}

func markSetupComplete() {
	setupComplete = true
}

func Init() {
	cfg := config.Get()
	if cfg == nil || len(cfg.Agents) < 1 {
		return
	}

	// Ensure primary agent is set
	_, exists := cfg.Agents[config.AgentPrimary]
	if exists {
		markSetupComplete()
	}
}

func CompleteSetup(provider models.ModelProvider, model models.Model, apiKey string) {
	err := config.Update(func(cfg *config.Config) {
		// Add Agent
		if cfg.Agents == nil {
			cfg.Agents = make(map[config.AgentName]config.Agent)
		}
		cfg.Agents[config.AgentPrimary] = config.Agent{
			Model:     model.ID,
			MaxTokens: model.DefaultMaxTokens,
		}
		cfg.Agents[config.AgentTitle] = config.Agent{
			Model:     model.ID,
			MaxTokens: 80,
		}
		cfg.Agents[config.AgentTask] = config.Agent{
			Model:     model.ID,
			MaxTokens: model.DefaultMaxTokens,
		}

		// Add Provider
		if cfg.Providers == nil {
			cfg.Providers = make(map[models.ModelProvider]config.Provider)
		}

		cfg.Providers[provider] = config.Provider{
			APIKey: apiKey,
		}
	})

	if err != nil {
		slog.Debug("Failed to complete setup", "error", err)
		panic(err)
	}

	markSetupComplete()
}
