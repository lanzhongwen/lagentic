package core

import (
	"context"
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
