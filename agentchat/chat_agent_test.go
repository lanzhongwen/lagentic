package agentchat

import (
	"context"
	"testing"

	"github.com/lanzhongwen/lagentic/core"
)

func TestResponse_Fields(t *testing.T) {
	msg := NewTextMessage("done", "coder")
	inner := []AgentEvent{
		{Type: "llm_call", Agent: "coder", Data: map[string]any{"model": "test"}},
	}
	resp := Response{ChatMessage: msg, InnerMessages: inner}
	if resp.ChatMessage.Type() != "TextMessage" {
		t.Errorf("ChatMessage.Type() = %q, want %q", resp.ChatMessage.Type(), "TextMessage")
	}
	if len(resp.InnerMessages) != 1 {
		t.Fatalf("len(InnerMessages) = %d, want 1", len(resp.InnerMessages))
	}
	if resp.InnerMessages[0].Type != "llm_call" {
		t.Errorf("InnerMessages[0].Type = %q, want %q", resp.InnerMessages[0].Type, "llm_call")
	}
}

func TestModelRef_String(t *testing.T) {
	ref := ModelRef{Provider: "anthropic", Model: "claude-sonnet-4-6"}
	want := "anthropic:claude-sonnet-4-6"
	if ref.String() != want {
		t.Errorf("ModelRef.String() = %q, want %q", ref.String(), want)
	}
}

func TestAgentEvent_Fields(t *testing.T) {
	evt := AgentEvent{
		Type:  "tool_call",
		Agent: "coder",
		Data:  map[string]any{"tool": "write_file"},
	}
	if evt.Type != "tool_call" {
		t.Errorf("Type = %q, want %q", evt.Type, "tool_call")
	}
	if evt.Agent != "coder" {
		t.Errorf("Agent = %q, want %q", evt.Agent, "coder")
	}
}

func TestChatAgent_Interface(t *testing.T) {
	var _ ChatAgent = &mockChatAgent{}
}

type mockChatAgent struct{}

func (m *mockChatAgent) Name() string { return "mock" }
func (m *mockChatAgent) Description() string { return "mock agent" }
func (m *mockChatAgent) OnMessages(_ context.Context, _ []ChatMessage, _ *core.CancellationToken) (Response, error) {
	return Response{ChatMessage: NewTextMessage("mock", "mock")}, nil
}
func (m *mockChatAgent) OnMessagesStream(_ context.Context, _ []ChatMessage, _ *core.CancellationToken) (<-chan AgentEvent, error) {
	ch := make(chan AgentEvent, 1)
	close(ch)
	return ch, nil
}
func (m *mockChatAgent) OnReset(_ context.Context) error { return nil }
func (m *mockChatAgent) SaveState() (map[string]any, error) { return nil, nil }
func (m *mockChatAgent) LoadState(_ map[string]any) error    { return nil }
func (m *mockChatAgent) Close() error                        { return nil }
