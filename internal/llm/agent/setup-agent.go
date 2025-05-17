package agent

import (
	"context"
	"fmt"
	"github.com/sst/opencode/internal/config"
	"github.com/sst/opencode/internal/llm/models"
	"github.com/sst/opencode/internal/llm/provider"
	"github.com/sst/opencode/internal/message"
)

type setupAgent struct {
}

func NewSetupAgent() (Service, error) {
	agent := &setupAgent{}

	return agent, nil
}

func (a *setupAgent) Cancel(sessionID string) {

}

func (a *setupAgent) IsBusy() bool {
	return true
}

func (a *setupAgent) IsSessionBusy(sessionID string) bool {
	return true
}

func (a *setupAgent) Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error) {
	return nil, ErrSessionBusy
}

func (a *setupAgent) GetUsage(ctx context.Context, sessionID string) (*int64, error) {
	usage := int64(0)

	return &usage, nil
}

func (a *setupAgent) EstimateContextWindowUsage(ctx context.Context, sessionID string) (float64, bool, error) {
	return 0, false, nil
}

func (a *setupAgent) TrackUsage(ctx context.Context, sessionID string, model models.Model, usage provider.TokenUsage) error {
	return nil
}

func (a *setupAgent) Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error) {
	return models.Model{}, fmt.Errorf("cannot change model while processing requests")
}

func (a *setupAgent) CompactSession(ctx context.Context, sessionID string, force bool) error {
	return nil
}
