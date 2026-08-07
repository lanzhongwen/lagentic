package agentchat

import (
	"context"
)

// LLMMessageRole constants for message roles.
type LLMMessageRole = string

const (
	LLMMessageRoleSystem    LLMMessageRole = "system"
	LLMMessageRoleUser      LLMMessageRole = "user"
	LLMMessageRoleAssistant LLMMessageRole = "assistant"
	LLMMessageRoleTool      LLMMessageRole = "tool"
)

// LLMMessage represents a single message in an LLM conversation.
type LLMMessage struct {
	Role       LLMMessageRole
	Content    string
	Name       string // optional: agent name or tool name
	ToolCallID string // optional: for tool result messages
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// LLMResponse is the result of a chat completion request.
type LLMResponse struct {
	Content      string
	FinishReason string
	ToolCalls    []ToolCall
	Usage        TokenUsage
}

// LLMStreamChunk is a single chunk in a streaming response.
type LLMStreamChunk struct {
	Content      string
	FinishReason string
	ToolCalls    []ToolCall
	Usage        TokenUsage
}

// ModelInfo describes an LLM model's capabilities and pricing.
type ModelInfo struct {
	Name        string
	MaxTokens   int
	InputPrice  float64 // per million tokens
	OutputPrice float64 // per million tokens
}

// CompletionOption configures a completion request.
type CompletionOption struct {
	MaxTokens   int
	Temperature float64
	StopWords   []string
}

// ChatCompletionClient abstracts LLM providers.
type ChatCompletionClient interface {
	Create(ctx context.Context, messages []LLMMessage, options ...CompletionOption) (LLMResponse, error)
	CreateStream(ctx context.Context, messages []LLMMessage, options ...CompletionOption) (<-chan LLMStreamChunk, error)
	ModelInfo() ModelInfo
}
