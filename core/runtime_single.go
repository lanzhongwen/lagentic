package core

import (
	"context"
	"fmt"
	"sync"
)

// SingleThreadedAgentRuntime processes messages synchronously and sequentially.
// No goroutines are spawned — all OnMessage calls happen in the caller's goroutine.
type SingleThreadedAgentRuntime struct {
	mu            sync.RWMutex
	agents        map[AgentID]Agent
	subscriptions []Subscription
}

// Compile-time interface check.
var _ AgentRuntime = (*SingleThreadedAgentRuntime)(nil)

// NewSingleThreadedAgentRuntime creates a new runtime.
func NewSingleThreadedAgentRuntime() *SingleThreadedAgentRuntime {
	return &SingleThreadedAgentRuntime{
		agents: make(map[AgentID]Agent),
	}
}

// RegisterFactory creates an agent via the factory and registers it.
// Any subscriptions provided are added to the runtime's subscription list.
func (r *SingleThreadedAgentRuntime) RegisterFactory(agentType string, factory AgentFactory, subs ...Subscription) error {
	agent, err := factory(r)
	if err != nil {
		return fmt.Errorf("factory for type %q: %w", agentType, err)
	}
	r.mu.Lock()
	r.agents[agent.ID()] = agent
	r.subscriptions = append(r.subscriptions, subs...)
	r.mu.Unlock()
	return nil
}

// RegisterAgent directly registers a pre-built agent.
func (r *SingleThreadedAgentRuntime) RegisterAgent(agent Agent) {
	r.mu.Lock()
	r.agents[agent.ID()] = agent
	r.mu.Unlock()
}

// AddSubscription adds a subscription to the runtime.
func (r *SingleThreadedAgentRuntime) AddSubscription(sub Subscription) error {
	r.mu.Lock()
	r.subscriptions = append(r.subscriptions, sub)
	r.mu.Unlock()
	return nil
}

// SendMessage delivers a message directly to a specific agent (RPC).
func (r *SingleThreadedAgentRuntime) SendMessage(ctx context.Context, msg any, recipient, sender AgentID) (any, error) {
	return r.SendMessageWithCancellationToken(ctx, msg, recipient, sender, nil)
}

// SendMessageWithCancellationToken delivers a message with cancellation support.
func (r *SingleThreadedAgentRuntime) SendMessageWithCancellationToken(ctx context.Context, msg any, recipient, sender AgentID, ct *CancellationToken) (any, error) {
	if ct != nil {
		select {
		case <-ct.Done():
			return nil, ErrContextCanceled
		default:
		}
	}

	r.mu.RLock()
	agent, ok := r.agents[recipient]
	r.mu.RUnlock()
	if !ok {
		return nil, ErrAgentNotFound
	}

	mc := MessageContext{
		IsRPC:             true,
		Sender:            sender,
		CancellationToken: ct,
	}
	return agent.OnMessage(ctx, msg, mc)
}

// PublishMessage delivers a message to all agents with matching subscriptions.
func (r *SingleThreadedAgentRuntime) PublishMessage(ctx context.Context, msg any, topic TopicID, sender AgentID) error {
	return r.PublishMessageWithCancellationToken(ctx, msg, topic, sender, nil)
}

// PublishMessageWithCancellationToken publishes with cancellation support.
func (r *SingleThreadedAgentRuntime) PublishMessageWithCancellationToken(ctx context.Context, msg any, topic TopicID, sender AgentID, ct *CancellationToken) error {
	if ct != nil {
		select {
		case <-ct.Done():
			return ErrContextCanceled
		default:
		}
	}

	r.mu.RLock()
	subs := make([]Subscription, len(r.subscriptions))
	copy(subs, r.subscriptions)
	r.mu.RUnlock()

	for _, sub := range subs {
		if !sub.IsMatch(topic) {
			continue
		}
		recipient := sub.MapToAgent(topic)

		r.mu.RLock()
		agent, ok := r.agents[recipient]
		r.mu.RUnlock()
		if !ok {
			continue // skip subscribers that aren't registered
		}

		mc := MessageContext{
			IsRPC:             false,
			TopicID:           topic,
			Sender:            sender,
			CancellationToken: ct,
		}
		if _, err := agent.OnMessage(ctx, msg, mc); err != nil {
			return fmt.Errorf("publish to %v: %w", recipient, err)
		}
	}
	return nil
}
