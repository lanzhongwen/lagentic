package core

import (
	"context"
	"errors"
	"testing"
)

// pingMsg and pongMsg are test message types for runtime tests.
type pingMsg struct {
	Text string
}

type pongMsg struct {
	Text string
}

// testAgent is a minimal Agent implementation for testing.
type testAgent struct {
	id          AgentID
	description string
	onMessage   func(ctx context.Context, msg any, mc MessageContext) (any, error)
}

func (a *testAgent) Metadata() AgentMetadata {
	return AgentMetadata{Type: a.id.Type, Description: a.description}
}
func (a *testAgent) ID() AgentID { return a.id }
func (a *testAgent) OnMessage(ctx context.Context, msg any, mc MessageContext) (any, error) {
	if a.onMessage != nil {
		return a.onMessage(ctx, msg, mc)
	}
	return nil, nil
}
func (a *testAgent) SaveState() (map[string]any, error) { return nil, nil }
func (a *testAgent) LoadState(_ map[string]any) error    { return nil }
func (a *testAgent) Close() error                        { return nil }

func TestSingleThreadedAgentRuntime_RegisterFactory(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()
	called := false
	err := rt.RegisterFactory("coder", func(runtime AgentRuntime) (Agent, error) {
		called = true
		return &testAgent{id: AgentID{Type: "coder", Key: "default"}, description: "test coder"}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFactory() error: %v", err)
	}
	if !called {
		t.Error("factory should have been called")
	}
}

func TestSingleThreadedAgentRuntime_RegisterFactory_Error(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()
	factoryErr := errors.New("factory failed")
	err := rt.RegisterFactory("bad", func(runtime AgentRuntime) (Agent, error) {
		return nil, factoryErr
	})
	if err == nil {
		t.Fatal("expected error from failing factory")
	}
	if !errors.Is(err, factoryErr) {
		t.Errorf("error = %v, want wrapped %v", err, factoryErr)
	}
}

func TestSingleThreadedAgentRuntime_SendMessage_DeliversToAgent(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	var received pingMsg
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, msg any, _ MessageContext) (any, error) {
			received = msg.(pingMsg)
			return pongMsg{Text: "pong"}, nil
		},
	}
	rt.RegisterAgent(agent)

	sender := AgentID{Type: "coordinator", Key: "default"}
	result, err := rt.SendMessage(context.Background(), pingMsg{Text: "ping"}, agent.ID(), sender)
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}
	pong, ok := result.(pongMsg)
	if !ok {
		t.Fatalf("result type = %T, want pongMsg", result)
	}
	if pong.Text != "pong" {
		t.Errorf("result.Text = %q, want %q", pong.Text, "pong")
	}
	if received.Text != "ping" {
		t.Errorf("received.Text = %q, want %q", received.Text, "ping")
	}
}

func TestSingleThreadedAgentRuntime_SendMessage_AgentNotFound(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()
	_, err := rt.SendMessage(context.Background(), "hello", AgentID{Type: "missing", Key: "x"}, AgentID{Type: "sender", Key: "x"})
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("error = %v, want ErrAgentNotFound", err)
	}
}

func TestSingleThreadedAgentRuntime_SendMessage_SetsMessageContext(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	var gotMC MessageContext
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, _ any, mc MessageContext) (any, error) {
			gotMC = mc
			return nil, nil
		},
	}
	rt.RegisterAgent(agent)

	sender := AgentID{Type: "coordinator", Key: "default"}
	_, err := rt.SendMessage(context.Background(), pingMsg{Text: "ping"}, agent.ID(), sender)
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}
	if !gotMC.IsRPC {
		t.Error("MessageContext.IsRPC should be true for SendMessage")
	}
	if gotMC.Sender != sender {
		t.Errorf("MessageContext.Sender = %v, want %v", gotMC.Sender, sender)
	}
}

func TestSingleThreadedAgentRuntime_PublishMessage_DeliversToSubscribers(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	var received pingMsg
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, msg any, _ MessageContext) (any, error) {
			received = msg.(pingMsg)
			return nil, nil
		},
	}
	rt.RegisterAgent(agent)
	rt.AddSubscription(NewTypeSubscription("task", "coder"))

	topic := TopicID{Type: "task", Source: "coordinator"}
	err := rt.PublishMessage(context.Background(), pingMsg{Text: "task-ready"}, topic, AgentID{Type: "coordinator", Key: "default"})
	if err != nil {
		t.Fatalf("PublishMessage() error: %v", err)
	}
	if received.Text != "task-ready" {
		t.Errorf("received.Text = %q, want %q", received.Text, "task-ready")
	}
}

func TestSingleThreadedAgentRuntime_PublishMessage_SetsMessageContext(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	var gotMC MessageContext
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, _ any, mc MessageContext) (any, error) {
			gotMC = mc
			return nil, nil
		},
	}
	rt.RegisterAgent(agent)
	rt.AddSubscription(NewTypeSubscription("task", "coder"))

	sender := AgentID{Type: "coordinator", Key: "default"}
	err := rt.PublishMessage(context.Background(), pingMsg{}, TopicID{Type: "task", Source: "coord"}, sender)
	if err != nil {
		t.Fatalf("PublishMessage() error: %v", err)
	}
	if gotMC.IsRPC {
		t.Error("MessageContext.IsRPC should be false for PublishMessage")
	}
	if gotMC.Sender != sender {
		t.Errorf("MessageContext.Sender = %v, want %v", gotMC.Sender, sender)
	}
}

func TestSingleThreadedAgentRuntime_PublishMessage_NoMatchingSubscription(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
	}
	rt.RegisterAgent(agent)
	// No subscription registered for "other" topic type.

	err := rt.PublishMessage(context.Background(), "hello", TopicID{Type: "other", Source: "x"}, AgentID{Type: "sender", Key: "x"})
	if err != nil {
		t.Fatalf("PublishMessage() with no matching sub should not error: %v", err)
	}
}

func TestSingleThreadedAgentRuntime_Cancellation(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	ct := NewCancellationToken()
	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, _ any, mc MessageContext) (any, error) {
			if mc.CancellationToken == nil {
				t.Error("CancellationToken should be set in MessageContext")
			}
			return nil, nil
		},
	}
	rt.RegisterAgent(agent)

	ctx := context.Background()
	// SendMessage with CancellationToken in context
	result, err := rt.SendMessageWithCancellationToken(ctx, pingMsg{Text: "ping"}, agent.ID(), AgentID{Type: "coordinator", Key: "default"}, ct)
	if err != nil {
		t.Fatalf("SendMessageWithCancellationToken() error: %v", err)
	}
	if result != nil {
		t.Errorf("result = %v, want nil", result)
	}
}

func TestSingleThreadedAgentRuntime_CanceledBeforeSend(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()
	ct := NewCancellationToken()
	ct.Cancel() // cancel immediately

	agent := &testAgent{
		id:          AgentID{Type: "coder", Key: "default"},
		description: "coder",
		onMessage: func(_ context.Context, _ any, _ MessageContext) (any, error) {
			t.Error("agent should not receive message after cancellation")
			return nil, nil
		},
	}
	rt.RegisterAgent(agent)

	_, err := rt.SendMessageWithCancellationToken(context.Background(), pingMsg{}, agent.ID(), AgentID{Type: "coord", Key: "default"}, ct)
	if !errors.Is(err, ErrContextCanceled) {
		t.Errorf("error = %v, want ErrContextCanceled", err)
	}
}

func TestSingleThreadedAgentRuntime_RegisterFactory_WithSubscription(t *testing.T) {
	rt := NewSingleThreadedAgentRuntime()

	var received pingMsg
	err := rt.RegisterFactory("coder", func(runtime AgentRuntime) (Agent, error) {
		return &testAgent{
			id:          AgentID{Type: "coder", Key: "default"},
			description: "coder",
			onMessage: func(_ context.Context, msg any, _ MessageContext) (any, error) {
				received = msg.(pingMsg)
				return nil, nil
			},
		}, nil
	}, NewTypeSubscription("task", "coder"))
	if err != nil {
		t.Fatalf("RegisterFactory() error: %v", err)
	}

	// Publish should reach the coder via the subscription registered with the factory.
	err = rt.PublishMessage(context.Background(), pingMsg{Text: "via-sub"}, TopicID{Type: "task", Source: "coord"}, AgentID{Type: "coord", Key: "default"})
	if err != nil {
		t.Fatalf("PublishMessage() error: %v", err)
	}
	if received.Text != "via-sub" {
		t.Errorf("received.Text = %q, want %q", received.Text, "via-sub")
	}
}
