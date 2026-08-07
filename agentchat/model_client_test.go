package agentchat

import (
	"context"
	"testing"
)

func TestLLMMessageRole_Values(t *testing.T) {
	if string(LLMMessageRoleSystem) != "system" {
		t.Errorf("LLMMessageRoleSystem = %q, want %q", string(LLMMessageRoleSystem), "system")
	}
	if string(LLMMessageRoleUser) != "user" {
		t.Errorf("LLMMessageRoleUser = %q, want %q", string(LLMMessageRoleUser), "user")
	}
	if string(LLMMessageRoleAssistant) != "assistant" {
		t.Errorf("LLMMessageRoleAssistant = %q, want %q", string(LLMMessageRoleAssistant), "assistant")
	}
	if string(LLMMessageRoleTool) != "tool" {
		t.Errorf("LLMMessageRoleTool = %q, want %q", string(LLMMessageRoleTool), "tool")
	}
}

func TestLLMMessage_Fields(t *testing.T) {
	msg := LLMMessage{
		Role:    LLMMessageRoleUser,
		Content: "write a function",
		Name:    "coordinator",
	}
	if msg.Role != LLMMessageRoleUser {
		t.Errorf("Role = %q, want %q", msg.Role, LLMMessageRoleUser)
	}
	if msg.Content != "write a function" {
		t.Errorf("Content = %q, want %q", msg.Content, "write a function")
	}
	if msg.Name != "coordinator" {
		t.Errorf("Name = %q, want %q", msg.Name, "coordinator")
	}
}

func TestLLMResponse_Fields(t *testing.T) {
	resp := LLMResponse{
		Content:      "here is the code",
		FinishReason: "stop",
		ToolCalls:    []ToolCall{{ID: "tc1", Name: "write_file", Arguments: `{}`}},
		Usage: TokenUsage{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}
	if resp.Content != "here is the code" {
		t.Errorf("Content = %q, want %q", resp.Content, "here is the code")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "stop")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "write_file" {
		t.Errorf("ToolCalls[0].Name = %q, want %q", resp.ToolCalls[0].Name, "write_file")
	}
	if resp.Usage.InputTokens != 100 {
		t.Errorf("Usage.InputTokens = %d, want 100", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 50 {
		t.Errorf("Usage.OutputTokens = %d, want 50", resp.Usage.OutputTokens)
	}
}

func TestModelInfo_Fields(t *testing.T) {
	info := ModelInfo{
		Name:        "claude-sonnet-4-6",
		MaxTokens:   8192,
		InputPrice:  3.0,
		OutputPrice: 15.0,
	}
	if info.Name != "claude-sonnet-4-6" {
		t.Errorf("Name = %q, want %q", info.Name, "claude-sonnet-4-6")
	}
	if info.MaxTokens != 8192 {
		t.Errorf("MaxTokens = %d, want 8192", info.MaxTokens)
	}
}

func TestChatCompletionClient_Interface(t *testing.T) {
	// Verify the interface exists by creating a mock implementation.
	var _ ChatCompletionClient = &mockClient{}
}

type mockClient struct{}

func (m *mockClient) Create(_ context.Context, _ []LLMMessage, _ ...CompletionOption) (LLMResponse, error) {
	return LLMResponse{Content: "mock", FinishReason: "stop"}, nil
}

func (m *mockClient) CreateStream(_ context.Context, _ []LLMMessage, _ ...CompletionOption) (<-chan LLMStreamChunk, error) {
	ch := make(chan LLMStreamChunk, 1)
	ch <- LLMStreamChunk{Content: "mock", FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (m *mockClient) ModelInfo() ModelInfo {
	return ModelInfo{Name: "mock", MaxTokens: 4096}
}
