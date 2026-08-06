package core

import (
	"context"
	"fmt"
)

// AgentMetadata describes an agent's type and purpose.
type AgentMetadata struct {
	Type        string
	Description string
}

// Agent is the minimal contract — mirrors AutoGen's Agent Protocol.
type Agent interface {
	Metadata() AgentMetadata
	ID() AgentID
	OnMessage(ctx context.Context, msg any, mc MessageContext) (any, error)
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
	Close() error
}

// AgentFactory creates an agent instance bound to a runtime.
type AgentFactory func(runtime AgentRuntime) (Agent, error)

// HandlerFunc handles a message of a specific type.
type HandlerFunc func(ctx context.Context, msg any, mc MessageContext) (any, error)

// AgentRuntime is the central orchestrator interface.
type AgentRuntime interface {
	SendMessage(ctx context.Context, msg any, recipient AgentID, sender AgentID) (any, error)
	PublishMessage(ctx context.Context, msg any, topic TopicID, sender AgentID) error
	RegisterFactory(agentType string, factory AgentFactory, subs ...Subscription) error
	AddSubscription(sub Subscription) error
}

// BaseAgent adds runtime binding and send/publish capabilities.
// SendMessage and PublishMessage fill in the sender AgentID automatically.
type BaseAgent struct {
	id          AgentID
	description string
	runtime     AgentRuntime
}

// NewBaseAgent creates a BaseAgent bound to the given runtime.
func NewBaseAgent(id AgentID, description string, runtime AgentRuntime) *BaseAgent {
	return &BaseAgent{id: id, description: description, runtime: runtime}
}

func (a *BaseAgent) ID() AgentID { return a.id }

func (a *BaseAgent) Metadata() AgentMetadata {
	return AgentMetadata{Type: a.id.Type, Description: a.description}
}

func (a *BaseAgent) OnMessage(_ context.Context, _ any, _ MessageContext) (any, error) {
	return nil, fmt.Errorf("BaseAgent.OnMessage: no handler registered, use RoutedAgent for message dispatch")
}

func (a *BaseAgent) SendMessage(ctx context.Context, msg any, recipient AgentID) (any, error) {
	return a.runtime.SendMessage(ctx, msg, recipient, a.id)
}

func (a *BaseAgent) PublishMessage(ctx context.Context, msg any, topic TopicID) error {
	return a.runtime.PublishMessage(ctx, msg, topic, a.id)
}

func (a *BaseAgent) SaveState() (map[string]any, error) { return nil, nil }

func (a *BaseAgent) LoadState(_ map[string]any) error { return nil }

func (a *BaseAgent) Close() error { return nil }
