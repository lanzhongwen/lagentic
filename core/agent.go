package core

import (
	"context"
	"fmt"
	"reflect"
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

// Compile-time checks that BaseAgent and RoutedAgent satisfy Agent.
var _ Agent = (*BaseAgent)(nil)
var _ Agent = (*RoutedAgent)(nil)

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
	return nil, fmt.Errorf("BaseAgent.OnMessage: %w", ErrNoHandler)
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

// handlerEntry pairs a handler with its dispatch mode.
type handlerEntry struct {
	isRPC   bool
	handler HandlerFunc
}

// RoutedAgent adds type-based message routing to BaseAgent.
// Handlers are registered via RegisterRPCHandler / RegisterEventHandler
// instead of overriding OnMessage directly.
type RoutedAgent struct {
	*BaseAgent
	handlers map[reflect.Type][]handlerEntry
}

// NewRoutedAgent creates a RoutedAgent bound to the given runtime.
func NewRoutedAgent(id AgentID, description string, runtime AgentRuntime) *RoutedAgent {
	return &RoutedAgent{
		BaseAgent: NewBaseAgent(id, description, runtime),
		handlers:  make(map[reflect.Type][]handlerEntry),
	}
}

// RegisterRPCHandler registers a handler for RPC messages of the given type.
// Only one RPC handler per message type is expected; the last registered wins.
func (a *RoutedAgent) RegisterRPCHandler(msgType any, handler HandlerFunc) {
	t := reflect.TypeOf(msgType)
	a.handlers[t] = append(a.handlers[t], handlerEntry{isRPC: true, handler: handler})
}

// RegisterEventHandler registers a handler for event messages of the given type.
// Multiple event handlers per type are called in registration order.
func (a *RoutedAgent) RegisterEventHandler(msgType any, handler HandlerFunc) {
	t := reflect.TypeOf(msgType)
	a.handlers[t] = append(a.handlers[t], handlerEntry{isRPC: false, handler: handler})
}

// OnMessage dispatches to registered handlers based on message type and IsRPC flag.
func (a *RoutedAgent) OnMessage(ctx context.Context, msg any, mc MessageContext) (any, error) {
	msgType := reflect.TypeOf(msg)
	entries, ok := a.handlers[msgType]
	if !ok {
		return nil, fmt.Errorf("no handler registered for message type %v: %w", msgType, ErrNoHandler)
	}

	var lastResult any
	var handlerCalled bool
	for _, entry := range entries {
		if mc.IsRPC && !entry.isRPC {
			continue
		}
		if !mc.IsRPC && entry.isRPC {
			continue
		}
		handlerCalled = true
		result, err := entry.handler(ctx, msg, mc)
		if err != nil {
			return nil, fmt.Errorf("handler for %v: %w", msgType, err)
		}
		lastResult = result
	}
	if !handlerCalled {
		return nil, fmt.Errorf("no %s handler registered for message type %v: %w", modeLabel(mc.IsRPC), msgType, ErrNoHandler)
	}
	return lastResult, nil
}

func modeLabel(isRPC bool) string {
	if isRPC {
		return "RPC"
	}
	return "event"
}
