package core

import (
	"context"
	"errors"
	"testing"
)

// mockRuntime is a test double for AgentRuntime.
type mockRuntime struct {
	sentMessages  []sentMessage
	publishedMsgs []publishedMessage
}

type sentMessage struct {
	ctx       context.Context
	msg       any
	recipient AgentID
	sender    AgentID
}

type publishedMessage struct {
	ctx    context.Context
	msg    any
	topic  TopicID
	sender AgentID
}

func (m *mockRuntime) SendMessage(ctx context.Context, msg any, recipient, sender AgentID) (any, error) {
	m.sentMessages = append(m.sentMessages, sentMessage{ctx, msg, recipient, sender})
	return "mock-response", nil
}

func (m *mockRuntime) PublishMessage(ctx context.Context, msg any, topic TopicID, sender AgentID) error {
	m.publishedMsgs = append(m.publishedMsgs, publishedMessage{ctx, msg, topic, sender})
	return nil
}

func (m *mockRuntime) RegisterFactory(agentType string, factory AgentFactory, subs ...Subscription) error {
	return nil
}

func (m *mockRuntime) AddSubscription(sub Subscription) error {
	return nil
}

func TestBaseAgent_ID(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "team-1"}
	agent := NewBaseAgent(id, "A coding agent", rt)
	if agent.ID() != id {
		t.Errorf("ID() = %v, want %v", agent.ID(), id)
	}
}

func TestBaseAgent_Metadata(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "team-1"}
	agent := NewBaseAgent(id, "A coding agent", rt)
	meta := agent.Metadata()
	if meta.Type != "coder" {
		t.Errorf("Metadata.Type = %q, want %q", meta.Type, "coder")
	}
	if meta.Description != "A coding agent" {
		t.Errorf("Metadata.Description = %q, want %q", meta.Description, "A coding agent")
	}
}

func TestBaseAgent_SendMessage_DelegatesToRuntime(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "team-1"}
	agent := NewBaseAgent(id, "coder", rt)

	recipient := AgentID{Type: "reviewer", Key: "team-1"}
	msg := "please review"
	ctx := context.Background()

	result, err := agent.SendMessage(ctx, msg, recipient)
	if err != nil {
		t.Fatalf("SendMessage() error: %v", err)
	}
	if result != "mock-response" {
		t.Errorf("SendMessage() = %v, want %q", result, "mock-response")
	}
	if len(rt.sentMessages) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(rt.sentMessages))
	}
	sent := rt.sentMessages[0]
	if sent.sender != id {
		t.Errorf("sender = %v, want %v", sent.sender, id)
	}
	if sent.recipient != recipient {
		t.Errorf("recipient = %v, want %v", sent.recipient, recipient)
	}
	if sent.msg != msg {
		t.Errorf("msg = %v, want %v", sent.msg, msg)
	}
}

func TestBaseAgent_PublishMessage_DelegatesToRuntime(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "team-1"}
	agent := NewBaseAgent(id, "coder", rt)

	topic := TopicID{Type: "task", Source: "coordinator"}
	msg := "code ready"
	ctx := context.Background()

	err := agent.PublishMessage(ctx, msg, topic)
	if err != nil {
		t.Fatalf("PublishMessage() error: %v", err)
	}
	if len(rt.publishedMsgs) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(rt.publishedMsgs))
	}
	pub := rt.publishedMsgs[0]
	if pub.sender != id {
		t.Errorf("sender = %v, want %v", pub.sender, id)
	}
	if pub.topic != topic {
		t.Errorf("topic = %v, want %v", pub.topic, topic)
	}
}

func TestBaseAgent_SaveLoadState(t *testing.T) {
	rt := &mockRuntime{}
	agent := NewBaseAgent(AgentID{Type: "coder", Key: "t1"}, "coder", rt)

	state, err := agent.SaveState()
	if err != nil {
		t.Fatalf("SaveState() error: %v", err)
	}
	if state != nil {
		t.Errorf("SaveState() = %v, want nil (default)", state)
	}

	err = agent.LoadState(nil)
	if err != nil {
		t.Fatalf("LoadState() error: %v", err)
	}
}

func TestBaseAgent_Close(t *testing.T) {
	rt := &mockRuntime{}
	agent := NewBaseAgent(AgentID{Type: "coder", Key: "t1"}, "coder", rt)
	if err := agent.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
}

func TestBaseAgent_OnMessage_ReturnsErrNoHandler(t *testing.T) {
	rt := &mockRuntime{}
	agent := NewBaseAgent(AgentID{Type: "coder", Key: "t1"}, "coder", rt)
	_, err := agent.OnMessage(context.Background(), "test", MessageContext{})
	if err == nil {
		t.Fatal("expected error from BaseAgent.OnMessage")
	}
	if !errors.Is(err, ErrNoHandler) {
		t.Errorf("errors.Is(err, ErrNoHandler) = false, err = %v", err)
	}
}

// testMsg is a concrete message type for testing RoutedAgent dispatch.
type testMsg struct {
	Content string
}

func TestRoutedAgent_RegisterRPCHandler_Dispatches(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	var received testMsg
	agent.RegisterRPCHandler(testMsg{}, func(_ context.Context, msg any, _ MessageContext) (any, error) {
		received = msg.(testMsg)
		return "handled", nil
	})

	ctx := context.Background()
	mc := MessageContext{IsRPC: true, Sender: AgentID{Type: "coordinator", Key: "t1"}}
	result, err := agent.OnMessage(ctx, testMsg{Content: "hello"}, mc)
	if err != nil {
		t.Fatalf("OnMessage() error: %v", err)
	}
	if result != "handled" {
		t.Errorf("result = %v, want %q", result, "handled")
	}
	if received.Content != "hello" {
		t.Errorf("received = %q, want %q", received.Content, "hello")
	}
}

func TestRoutedAgent_RegisterEventHandler_Dispatches(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	var called bool
	agent.RegisterEventHandler(testMsg{}, func(_ context.Context, _ any, _ MessageContext) (any, error) {
		called = true
		return nil, nil
	})

	ctx := context.Background()
	mc := MessageContext{IsRPC: false, Sender: AgentID{Type: "coordinator", Key: "t1"}}
	_, err := agent.OnMessage(ctx, testMsg{Content: "event"}, mc)
	if err != nil {
		t.Fatalf("OnMessage() error: %v", err)
	}
	if !called {
		t.Error("event handler should have been called")
	}
}

func TestRoutedAgent_OnMessage_UnregisteredType(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	ctx := context.Background()
	mc := MessageContext{IsRPC: true}
	_, err := agent.OnMessage(ctx, testMsg{Content: "hello"}, mc)
	if err == nil {
		t.Error("expected error for unregistered message type")
	}
	if !errors.Is(err, ErrNoHandler) {
		t.Errorf("errors.Is(err, ErrNoHandler) = false, err = %v", err)
	}
}

func TestRoutedAgent_OnMessage_NoMatchingModeHandler(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	// Register only an event handler
	agent.RegisterEventHandler(testMsg{}, func(_ context.Context, _ any, _ MessageContext) (any, error) {
		return "event-handled", nil
	})

	// Send an RPC message — should get ErrNoHandler
	ctx := context.Background()
	mc := MessageContext{IsRPC: true}
	_, err := agent.OnMessage(ctx, testMsg{Content: "hello"}, mc)
	if !errors.Is(err, ErrNoHandler) {
		t.Errorf("error = %v, want ErrNoHandler", err)
	}
}

func TestRoutedAgent_MultipleEventHandlers(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	var callOrder []int
	agent.RegisterEventHandler(testMsg{}, func(_ context.Context, _ any, _ MessageContext) (any, error) {
		callOrder = append(callOrder, 1)
		return nil, nil
	})
	agent.RegisterEventHandler(testMsg{}, func(_ context.Context, _ any, _ MessageContext) (any, error) {
		callOrder = append(callOrder, 2)
		return nil, nil
	})

	ctx := context.Background()
	mc := MessageContext{IsRPC: false}
	_, err := agent.OnMessage(ctx, testMsg{Content: "event"}, mc)
	if err != nil {
		t.Fatalf("OnMessage() error: %v", err)
	}
	if len(callOrder) != 2 {
		t.Fatalf("expected 2 handler calls, got %d", len(callOrder))
	}
	if callOrder[0] != 1 || callOrder[1] != 2 {
		t.Errorf("call order = %v, want [1 2]", callOrder)
	}
}

func TestRoutedAgent_RPC_IgnoresEventHandlers(t *testing.T) {
	rt := &mockRuntime{}
	id := AgentID{Type: "coder", Key: "t1"}
	agent := NewRoutedAgent(id, "coder", rt)

	var eventCalled bool
	agent.RegisterEventHandler(testMsg{}, func(_ context.Context, _ any, _ MessageContext) (any, error) {
		eventCalled = true
		return nil, nil
	})
	agent.RegisterRPCHandler(testMsg{}, func(_ context.Context, msg any, _ MessageContext) (any, error) {
		return "rpc-handled", nil
	})

	ctx := context.Background()
	mc := MessageContext{IsRPC: true}
	result, err := agent.OnMessage(ctx, testMsg{Content: "rpc"}, mc)
	if err != nil {
		t.Fatalf("OnMessage() error: %v", err)
	}
	if result != "rpc-handled" {
		t.Errorf("result = %v, want %q", result, "rpc-handled")
	}
	if eventCalled {
		t.Error("event handler should not be called for RPC messages")
	}
}
