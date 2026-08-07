package agentchat

import (
	"context"
	"fmt"

	"github.com/lanzhongwen/lagentic/core"
)

// ChatMessage is an alias for BaseChatMessage used in ChatAgent method signatures.
type ChatMessage = BaseChatMessage

// ChatAgent is the high-level agent interface for LLM-aware agents.
type ChatAgent interface {
	Name() string
	Description() string
	OnMessages(ctx context.Context, messages []ChatMessage, ct *core.CancellationToken) (Response, error)
	OnMessagesStream(ctx context.Context, messages []ChatMessage, ct *core.CancellationToken) (<-chan AgentEvent, error)
	OnReset(ctx context.Context) error
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
	Close() error
}

// Response from a ChatAgent.
type Response struct {
	ChatMessage   ChatMessage
	InnerMessages []AgentEvent
}

// ModelRef identifies a provider + model combination (e.g., "anthropic:claude-sonnet-4-6").
type ModelRef struct {
	Provider string
	Model    string
}

func (r ModelRef) String() string {
	return fmt.Sprintf("%s:%s", r.Provider, r.Model)
}

// AgentEvent represents an observable event during agent execution.
type AgentEvent struct {
	Type  string
	Agent string
	Data  any
}
