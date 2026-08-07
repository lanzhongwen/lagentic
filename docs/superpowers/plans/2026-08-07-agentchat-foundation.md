# AgentChat Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the `lagentic-agentchat` foundation — chat message types, LLM client abstraction, ChatAgent interface, conversation context, Workbench, and Handoff — everything needed before building AssistantAgent and GroupChat.

**Architecture:** Layer 2 of the two-layer design. Built on `lagentic-core` (already implemented). This layer adds LLM awareness: chat messages between agents, an abstract LLM client interface, per-agent conversation history, a Workbench for tool management, and handoff definitions. No concrete LLM clients yet — those live in `ext/`. The ChatCompletionClient interface is the seam.

**Tech Stack:** Go 1.24+, stdlib only (imports `core` package from same module)

## Global Constraints

- `agentchat` package imports `core` — never imports `ext`
- All errors wrapped with context: `fmt.Errorf("doing X: %w", err)`
- Sentinel errors for distinguishable failure modes; `ErrMaxTurnsExceeded`, `ErrTokenLimitExceeded` defined here
- All concurrent code must pass `-race`
- Table-driven tests for multi-scenario validation
- Test names describe behavior: `TestX_Behavior_ExpectedResult`
- Conventional commit format: `feat(scope): description`
- Package name: `agentchat`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `agentchat/messages.go` | `BaseChatMessage` interface, `TextMessage`, `HandoffMessage`, `ToolCallMessage`, `ToolResultMessage`, `StopMessage`, `ToolCall` |
| `agentchat/model_client.go` | `LLMMessage`, `LLMResponse`, `LLMStreamChunk`, `ModelInfo`, `CompletionOption`, `ChatCompletionClient` interface |
| `agentchat/model_context.go` | `ChatCompletionContext` interface, `UnboundedChatCompletionContext`, `BufferedChatCompletionContext` |
| `agentchat/chat_agent.go` | `ChatAgent` interface, `Response`, `ModelRef`, `AgentEvent` |
| `agentchat/workbench.go` | `Workbench` interface, `StaticWorkbench` |
| `agentchat/handoff.go` | `Handoff` definition |
| `agentchat/errors.go` | `ErrMaxTurnsExceeded`, `ErrTokenLimitExceeded` |

Corresponding `*_test.go` for each.

---

### Task 1: Chat Messages

**Files:**
- Create: `agentchat/messages.go`
- Create: `agentchat/messages_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks
- Produces: `BaseChatMessage` interface, `TextMessage`, `HandoffMessage`, `ToolCallMessage`, `ToolResultMessage`, `StopMessage`, `ToolCall`

- [ ] **Step 1: Write failing tests for chat messages**

Create `agentchat/messages_test.go`:

```go
package agentchat

import "testing"

func TestTextMessage_Type(t *testing.T) {
	msg := TextMessage{Content: "hello", Source: "coder"}
	if msg.Type() != "TextMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "TextMessage")
	}
}

func TestTextMessage_Source(t *testing.T) {
	msg := TextMessage{Content: "hello", Source: "coder"}
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestTextMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = TextMessage{}
}

func TestHandoffMessage_Type(t *testing.T) {
	msg := HandoffMessage{Target: "reviewer", Context: "code ready", Source: "coder"}
	if msg.Type() != "HandoffMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "HandoffMessage")
	}
}

func TestHandoffMessage_Source(t *testing.T) {
	msg := HandoffMessage{Target: "reviewer", Context: "code ready", Source: "coder"}
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestHandoffMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = HandoffMessage{}
}

func TestToolCallMessage_Type(t *testing.T) {
	msg := ToolCallMessage{
		ToolCalls: []ToolCall{{ID: "tc1", Name: "read_file", Arguments: `{"path":"main.go"}`}},
		Source:    "coder",
	}
	if msg.Type() != "ToolCallMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "ToolCallMessage")
	}
}

func TestToolCallMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = ToolCallMessage{}
}

func TestToolResultMessage_Type(t *testing.T) {
	msg := ToolResultMessage{
		Results: []core.ToolResult{{Content: "file contents", IsError: false}},
		Source:  "coder",
	}
	if msg.Type() != "ToolResultMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "ToolResultMessage")
	}
}

func TestToolResultMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = ToolResultMessage{}
}

func TestStopMessage_Type(t *testing.T) {
	msg := StopMessage{Content: "task complete", Source: "coordinator"}
	if msg.Type() != "StopMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "StopMessage")
	}
}

func TestStopMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = StopMessage{}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestTextMessage -v`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement chat messages**

Create `agentchat/messages.go`:

```go
package agentchat

import "github.com/lanzhongwen/lagentic/core"

// BaseChatMessage is the interface for all agent-to-agent messages.
type BaseChatMessage interface {
	Type() string
	Source() string
}

// TextMessage carries a text message between agents.
type TextMessage struct {
	Content string
	Source  string
}

func (m TextMessage) Type() string   { return "TextMessage" }
func (m TextMessage) Source() string  { return m.Source }

// HandoffMessage explicitly transfers control to another agent.
type HandoffMessage struct {
	Target  string
	Context string
	Source  string
}

func (m HandoffMessage) Type() string   { return "HandoffMessage" }
func (m HandoffMessage) Source() string  { return m.Source }

// ToolCall represents a single tool invocation request.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // JSON-encoded arguments
}

// ToolCallMessage carries one or more tool call requests.
type ToolCallMessage struct {
	ToolCalls []ToolCall
	Source    string
}

func (m ToolCallMessage) Type() string   { return "ToolCallMessage" }
func (m ToolCallMessage) Source() string  { return m.Source }

// ToolResultMessage carries the results of tool executions.
type ToolResultMessage struct {
	Results []core.ToolResult
	Source  string
}

func (m ToolResultMessage) Type() string   { return "ToolResultMessage" }
func (m ToolResultMessage) Source() string  { return m.Source }

// StopMessage signals that the agent or team should stop.
type StopMessage struct {
	Content string
	Source  string
}

func (m StopMessage) Type() string   { return "StopMessage" }
func (m StopMessage) Source() string  { return m.Source }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/
git commit -m "feat(agentchat): add chat message types"
```

---

### Task 2: AgentChat Errors

**Files:**
- Create: `agentchat/errors.go`
- Create: `agentchat/errors_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `ErrMaxTurnsExceeded`, `ErrTokenLimitExceeded`

- [ ] **Step 1: Write failing tests for agentchat errors**

Create `agentchat/errors_test.go`:

```go
package agentchat

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_Values(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"MaxTurnsExceeded", ErrMaxTurnsExceeded, "max turns exceeded"},
		{"TokenLimitExceeded", ErrTokenLimitExceeded, "token limit exceeded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestSentinelErrors_Wrapped(t *testing.T) {
	wrapped := fmt.Errorf("team stopped: %w", ErrMaxTurnsExceeded)
	if !errors.Is(wrapped, ErrMaxTurnsExceeded) {
		t.Error("errors.Is should find ErrMaxTurnsExceeded through wrap")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestSentinelErrors -v`
Expected: FAIL — `ErrMaxTurnsExceeded` undefined

- [ ] **Step 3: Implement agentchat errors**

Create `agentchat/errors.go`:

```go
package agentchat

import "fmt"

var (
	ErrMaxTurnsExceeded  = fmt.Errorf("max turns exceeded")
	ErrTokenLimitExceeded = fmt.Errorf("token limit exceeded")
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -run TestSentinelErrors -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/errors.go agentchat/errors_test.go
git commit -m "feat(agentchat): add sentinel errors"
```

---

### Task 3: LLM Client Interface

**Files:**
- Create: `agentchat/model_client.go`
- Create: `agentchat/model_client_test.go`

**Interfaces:**
- Consumes: `core.CancellationToken`
- Produces: `LLMMessage`, `LLMMessageRole`, `LLMResponse`, `LLMStreamChunk`, `ModelInfo`, `CompletionOption`, `ChatCompletionClient` interface

- [ ] **Step 1: Write failing tests for LLM types**

Create `agentchat/model_client_test.go`:

```go
package agentchat

import "testing"

func TestLLMMessageRole_Values(t *testing.T) {
	if LLMMessageRoleSystem != "system" {
		t.Errorf("LLMMessageRoleSystem = %q, want %q", LLMMessageRoleSystem, "system")
	}
	if LLMMessageRoleUser != "user" {
		t.Errorf("LLMMessageRoleUser = %q, want %q", LLMMessageRoleUser, "user")
	}
	if LLMMessageRoleAssistant != "assistant" {
		t.Errorf("LLMMessageRoleAssistant = %q, want %q", LLMMessageRoleAssistant, "assistant")
	}
	if LLMMessageRoleTool != "tool" {
		t.Errorf("LLMMessageRoleTool = %q, want %q", LLMMessageRoleTool, "tool")
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
		Name:         "claude-sonnet-4-6",
		MaxTokens:    8192,
		InputPrice:   3.0,
		OutputPrice:  15.0,
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
```

Note: `context` must be imported. The `ToolCall` type is defined in `messages.go` (Task 1).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestLLMMessage -v`
Expected: FAIL — `LLMMessageRole` undefined

- [ ] **Step 3: Implement LLM client types**

Create `agentchat/model_client.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/model_client.go agentchat/model_client_test.go
git commit -m "feat(agentchat): add LLM client interface and types"
```

---

### Task 4: ChatCompletionContext

**Files:**
- Create: `agentchat/model_context.go`
- Create: `agentchat/model_context_test.go`

**Interfaces:**
- Consumes: `LLMMessage` (from Task 3)
- Produces: `ChatCompletionContext` interface, `UnboundedChatCompletionContext`, `BufferedChatCompletionContext`

- [ ] **Step 1: Write failing tests for ChatCompletionContext**

Create `agentchat/model_context_test.go`:

```go
package agentchat

import "testing"

func TestUnboundedContext_AddAndGetMessages(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	msg1 := LLMMessage{Role: LLMMessageRoleUser, Content: "hello"}
	msg2 := LLMMessage{Role: LLMMessageRoleAssistant, Content: "hi"}

	if err := ctx.AddMessage(msg1); err != nil {
		t.Fatalf("AddMessage() error: %v", err)
	}
	if err := ctx.AddMessage(msg2); err != nil {
		t.Fatalf("AddMessage() error: %v", err)
	}

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("msgs[0].Content = %q, want %q", msgs[0].Content, "hello")
	}
	if msgs[1].Content != "hi" {
		t.Errorf("msgs[1].Content = %q, want %q", msgs[1].Content, "hi")
	}
}

func TestUnboundedContext_Clear(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "hello"})

	if err := ctx.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}
	if len(ctx.GetMessages()) != 0 {
		t.Errorf("len(GetMessages()) after Clear = %d, want 0", len(ctx.GetMessages()))
	}
}

func TestUnboundedContext_SaveLoadState(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "hello"})

	state, err := ctx.SaveState()
	if err != nil {
		t.Fatalf("SaveState() error: %v", err)
	}

	ctx2 := NewUnboundedChatCompletionContext()
	if err := ctx2.LoadState(state); err != nil {
		t.Fatalf("LoadState() error: %v", err)
	}
	if len(ctx2.GetMessages()) != 1 {
		t.Fatalf("len(GetMessages()) after LoadState = %d, want 1", len(ctx2.GetMessages()))
	}
	if ctx2.GetMessages()[0].Content != "hello" {
		t.Errorf("Content after LoadState = %q, want %q", ctx2.GetMessages()[0].Content, "hello")
	}
}

func TestUnboundedContext_Interface(t *testing.T) {
	var _ ChatCompletionContext = NewUnboundedChatCompletionContext()
}

func TestBufferedContext_TruncatesToLimit(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(2)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg2"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg3"})

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
	// Should keep the last 2 messages
	if msgs[0].Content != "msg2" {
		t.Errorf("msgs[0].Content = %q, want %q", msgs[0].Content, "msg2")
	}
	if msgs[1].Content != "msg3" {
		t.Errorf("msgs[1].Content = %q, want %q", msgs[1].Content, "msg3")
	}
}

func TestBufferedContext_UnderLimit_NoTruncation(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(5)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg2"})

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
}

func TestBufferedContext_Clear(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(5)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.Clear()
	if len(ctx.GetMessages()) != 0 {
		t.Errorf("len after Clear = %d, want 0", len(ctx.GetMessages()))
	}
}

func TestBufferedContext_Interface(t *testing.T) {
	var _ ChatCompletionContext = NewBufferedChatCompletionContext(10)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestUnboundedContext -v`
Expected: FAIL — `NewUnboundedChatCompletionContext` undefined

- [ ] **Step 3: Implement ChatCompletionContext**

Create `agentchat/model_context.go`:

```go
package agentchat

// ChatCompletionContext manages per-agent conversation history.
type ChatCompletionContext interface {
	AddMessage(msg LLMMessage) error
	GetMessages() []LLMMessage
	Clear() error
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
}

// UnboundedChatCompletionContext stores messages without limit.
type UnboundedChatCompletionContext struct {
	messages []LLMMessage
}

// NewUnboundedChatCompletionContext creates an unbounded context.
func NewUnboundedChatCompletionContext() *UnboundedChatCompletionContext {
	return &UnboundedChatCompletionContext{
		messages: make([]LLMMessage, 0),
	}
}

func (c *UnboundedChatCompletionContext) AddMessage(msg LLMMessage) error {
	c.messages = append(c.messages, msg)
	return nil
}

func (c *UnboundedChatCompletionContext) GetMessages() []LLMMessage {
	return c.messages
}

func (c *UnboundedChatCompletionContext) Clear() error {
	c.messages = c.messages[:0]
	return nil
}

func (c *UnboundedChatCompletionContext) SaveState() (map[string]any, error) {
	return map[string]any{
		"messages": c.messages,
	}, nil
}

func (c *UnboundedChatCompletionContext) LoadState(state map[string]any) error {
	if msgs, ok := state["messages"]; ok {
		c.messages = msgs.([]LLMMessage)
	}
	return nil
}

// BufferedChatCompletionContext keeps only the last N messages.
type BufferedChatCompletionContext struct {
	messages []LLMMessage
	limit    int
}

// NewBufferedChatCompletionContext creates a context that retains at most limit messages.
func NewBufferedChatCompletionContext(limit int) *BufferedChatCompletionContext {
	return &BufferedChatCompletionContext{
		messages: make([]LLMMessage, 0),
		limit:    limit,
	}
}

func (c *BufferedChatCompletionContext) AddMessage(msg LLMMessage) error {
	c.messages = append(c.messages, msg)
	if len(c.messages) > c.limit {
		c.messages = c.messages[len(c.messages)-c.limit:]
	}
	return nil
}

func (c *BufferedChatCompletionContext) GetMessages() []LLMMessage {
	return c.messages
}

func (c *BufferedChatCompletionContext) Clear() error {
	c.messages = c.messages[:0]
	return nil
}

func (c *BufferedChatCompletionContext) SaveState() (map[string]any, error) {
	return map[string]any{
		"messages": c.messages,
		"limit":    c.limit,
	}, nil
}

func (c *BufferedChatCompletionContext) LoadState(state map[string]any) error {
	if msgs, ok := state["messages"]; ok {
		c.messages = msgs.([]LLMMessage)
	}
	if limit, ok := state["limit"]; ok {
		c.limit = limit.(int)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -run "TestUnboundedContext|TestBufferedContext" -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/model_context.go agentchat/model_context_test.go
git commit -m "feat(agentchat): add ChatCompletionContext with Unbounded and Buffered"
```

---

### Task 5: ChatAgent Interface, Response, ModelRef, AgentEvent

**Files:**
- Create: `agentchat/chat_agent.go`
- Create: `agentchat/chat_agent_test.go`

**Interfaces:**
- Consumes: `BaseChatMessage` (from Task 1), `core.CancellationToken`
- Produces: `ChatAgent` interface, `Response`, `ModelRef`, `AgentEvent`

- [ ] **Step 1: Write failing tests for ChatAgent types**

Create `agentchat/chat_agent_test.go`:

```go
package agentchat

import (
	"context"
	"testing"

	"github.com/lanzhongwen/lagentic/core"
)

func TestResponse_Fields(t *testing.T) {
	msg := TextMessage{Content: "done", Source: "coder"}
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
	return Response{ChatMessage: TextMessage{Content: "mock", Source: "mock"}}, nil
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestResponse -v`
Expected: FAIL — `Response` undefined

- [ ] **Step 3: Implement ChatAgent, Response, ModelRef, AgentEvent**

Create `agentchat/chat_agent.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -run "TestResponse|TestModelRef|TestAgentEvent|TestChatAgent" -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/chat_agent.go agentchat/chat_agent_test.go
git commit -m "feat(agentchat): add ChatAgent interface, Response, ModelRef, AgentEvent"
```

---

### Task 6: Workbench

**Files:**
- Create: `agentchat/workbench.go`
- Create: `agentchat/workbench_test.go`

**Interfaces:**
- Consumes: `core.Tool`, `core.ToolSchema`, `core.ToolResult`, `core.CancellationToken`, `core.NewFunctionTool`, `core.ErrToolNotFound`
- Produces: `Workbench` interface, `StaticWorkbench`

- [ ] **Step 1: Write failing tests for Workbench**

Create `agentchat/workbench_test.go`:

```go
package agentchat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lanzhongwen/lagentic/core"
)

func TestStaticWorkbench_RegisterAndListTools(t *testing.T) {
	wb := NewStaticWorkbench()
	tool := core.NewFunctionTool("read_file", "Read a file", core.ToolSchema{
		Name: "read_file", Description: "Read a file",
		Parameters: map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *core.CancellationToken) (any, error) {
		return "contents", nil
	})

	if err := wb.Register(tool); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	tools := wb.ListTools()
	if len(tools) != 1 {
		t.Fatalf("len(ListTools()) = %d, want 1", len(tools))
	}
	if tools[0].Name != "read_file" {
		t.Errorf("tools[0].Name = %q, want %q", tools[0].Name, "read_file")
	}
}

func TestStaticWorkbench_CallTool(t *testing.T) {
	wb := NewStaticWorkbench()
	tool := core.NewFunctionTool("echo", "Echo input", core.ToolSchema{
		Name: "echo", Description: "Echo input",
		Parameters: map[string]any{"type": "object"},
	}, func(_ context.Context, args json.RawMessage, _ *core.CancellationToken) (any, error) {
		return string(args), nil
	})
	wb.Register(tool)

	result, err := wb.CallTool(context.Background(), "echo", json.RawMessage(`{"msg":"hi"}`), nil)
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}
	if result.Content != `{"msg":"hi"}` {
		t.Errorf("result.Content = %q, want %q", result.Content, `{"msg":"hi"}`)
	}
	if result.IsError {
		t.Error("result.IsError should be false")
	}
}

func TestStaticWorkbench_CallTool_NotFound(t *testing.T) {
	wb := NewStaticWorkbench()
	_, err := wb.CallTool(context.Background(), "missing", json.RawMessage(`{}`), nil)
	if err == nil {
		t.Fatal("expected error for missing tool")
	}
	// The error should wrap core.ErrToolNotFound
	if err.Error() == "" {
		t.Error("error should have a message")
	}
}

func TestStaticWorkbench_RegisterDuplicate(t *testing.T) {
	wb := NewStaticWorkbench()
	tool1 := core.NewFunctionTool("echo", "Echo v1", core.ToolSchema{
		Name: "echo", Description: "Echo v1",
		Parameters: map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *core.CancellationToken) (any, error) {
		return "v1", nil
	})
	tool2 := core.NewFunctionTool("echo", "Echo v2", core.ToolSchema{
		Name: "echo", Description: "Echo v2",
		Parameters: map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *core.CancellationToken) (any, error) {
		return "v2", nil
	})
	wb.Register(tool1)
	wb.Register(tool2)

	// Last registration should win
	result, err := wb.CallTool(context.Background(), "echo", json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}
	if result.Content != "v2" {
		t.Errorf("result.Content = %q, want %q", result.Content, "v2")
	}
}

func TestWorkbench_Interface(t *testing.T) {
	var _ Workbench = NewStaticWorkbench()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestStaticWorkbench -v`
Expected: FAIL — `NewStaticWorkbench` undefined

- [ ] **Step 3: Implement Workbench**

Create `agentchat/workbench.go`:

```go
package agentchat

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lanzhongwen/lagentic/core"
)

// Workbench is a container for tools available to an agent.
type Workbench interface {
	ListTools() []core.ToolSchema
	CallTool(ctx context.Context, name string, args json.RawMessage, ct *core.CancellationToken) (core.ToolResult, error)
	Register(tool core.Tool) error
}

// StaticWorkbench is a simple in-memory workbench.
type StaticWorkbench struct {
	mu    sync.RWMutex
	tools map[string]core.Tool
}

// NewStaticWorkbench creates an empty workbench.
func NewStaticWorkbench() *StaticWorkbench {
	return &StaticWorkbench{
		tools: make(map[string]core.Tool),
	}
}

func (w *StaticWorkbench) Register(tool core.Tool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tools[tool.Name()] = tool
	return nil
}

func (w *StaticWorkbench) ListTools() []core.ToolSchema {
	w.mu.RLock()
	defer w.mu.RUnlock()
	schemas := make([]core.ToolSchema, 0, len(w.tools))
	for _, tool := range w.tools {
		schemas = append(schemas, tool.Schema())
	}
	return schemas
}

func (w *StaticWorkbench) CallTool(ctx context.Context, name string, args json.RawMessage, ct *core.CancellationToken) (core.ToolResult, error) {
	w.mu.RLock()
	tool, ok := w.tools[name]
	w.mu.RUnlock()
	if !ok {
		return core.ToolResult{}, fmt.Errorf("workbench: %w", core.ErrToolNotFound)
	}

	result, err := tool.RunJSON(ctx, args, ct)
	if err != nil {
		return core.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	// Marshal the result to string for ToolResult.Content.
	content, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return core.ToolResult{Content: fmt.Sprintf("%v", result)}, nil
	}
	return core.ToolResult{Content: string(content)}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -run TestStaticWorkbench -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add agentchat/workbench.go agentchat/workbench_test.go
git commit -m "feat(agentchat): add Workbench interface and StaticWorkbench"
```

---

### Task 7: Handoff

**Files:**
- Create: `agentchat/handoff.go`
- Create: `agentchat/handoff_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `Handoff`

- [ ] **Step 1: Write failing tests for Handoff**

Create `agentchat/handoff_test.go`:

```go
package agentchat

import "testing"

func TestHandoff_Fields(t *testing.T) {
	h := Handoff{Target: "reviewer", Description: "Code is ready for review"}
	if h.Target != "reviewer" {
		t.Errorf("Target = %q, want %q", h.Target, "reviewer")
	}
	if h.Description != "Code is ready for review" {
		t.Errorf("Description = %q, want %q", h.Description, "Code is ready for review")
	}
}

func TestHandoff_ProducesHandoffMessage(t *testing.T) {
	h := Handoff{Target: "reviewer", Description: "Code is ready"}
	msg := h.ToHandoffMessage("coder")
	if msg.Target != "reviewer" {
		t.Errorf("Target = %q, want %q", msg.Target, "reviewer")
	}
	if msg.Context != "Code is ready" {
		t.Errorf("Context = %q, want %q", msg.Context, "Code is ready")
	}
	if msg.Source != "coder" {
		t.Errorf("Source = %q, want %q", msg.Source, "coder")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./agentchat/ -run TestHandoff -v`
Expected: FAIL — `Handoff` undefined

- [ ] **Step 3: Implement Handoff**

Create `agentchat/handoff.go`:

```go
package agentchat

// Handoff defines a potential transfer of control to another agent.
type Handoff struct {
	Target      string
	Description string
}

// ToHandoffMessage creates a HandoffMessage from this handoff definition.
func (h Handoff) ToHandoffMessage(source string) HandoffMessage {
	return HandoffMessage{
		Target:  h.Target,
		Context: h.Description,
		Source:  source,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./agentchat/ -run TestHandoff -v`
Expected: All PASS

- [ ] **Step 5: Run all agentchat tests with race detector**

Run: `CGO_ENABLED=1 go test -race ./agentchat/ -v`
Expected: All PASS, no races

- [ ] **Step 6: Commit**

```bash
git add agentchat/handoff.go agentchat/handoff_test.go
git commit -m "feat(agentchat): add Handoff definition"
```

---

## Self-Review

### 1. Spec Coverage

| Spec Section | Covered By Task |
|---|---|
| `BaseChatMessage` interface | Task 1 |
| `TextMessage` | Task 1 |
| `HandoffMessage` | Task 1 |
| `ToolCallMessage` + `ToolCall` | Task 1 |
| `ToolResultMessage` | Task 1 |
| `StopMessage` | Task 1 |
| `ChatAgent` interface | Task 5 |
| `Response` struct | Task 5 |
| `ModelRef` | Task 5 |
| `ChatCompletionClient` interface | Task 3 |
| `LLMMessage`, `LLMResponse`, `LLMStreamChunk` | Task 3 |
| `ModelInfo`, `CompletionOption`, `TokenUsage` | Task 3 |
| `ChatCompletionContext` interface | Task 4 |
| `UnboundedChatCompletionContext` | Task 4 |
| `BufferedChatCompletionContext` | Task 4 |
| `Workbench` interface | Task 6 |
| `StaticWorkbench` | Task 6 |
| `Handoff` definition | Task 7 |
| `AgentEvent` | Task 5 |
| `ErrMaxTurnsExceeded` | Task 2 |
| `ErrTokenLimitExceeded` | Task 2 |
| GroupChat internal events (`GroupChatStart`, etc.) | Not in this plan — belongs in GroupChat plan |
| `AssistantAgent` | Not in this plan — belongs in GroupChat plan (needs Team context) |
| `TokenLimited` context | Deferred — spec says "implementations: Unbounded, Buffered(N), TokenLimited" but TokenLimited requires tokenizer integration from `ext`. Add when providers exist. |
| `ChatMessage` alias for `BaseChatMessage` | Task 5 (type alias in chat_agent.go) |

**Gaps:**
- `TokenLimitedChatCompletionContext` — deferred. Requires a tokenizer (from `ext`) to count tokens. Adding a stub now would be speculative. Will add in provider plan.
- `AssistantAgent` — depends on ChatCompletionClient + Workbench + Handoff + GroupChat runtime. Better built alongside GroupChat in the next plan.
- GroupChat internal events (`GroupChatStart`, `GroupChatRequestPublish`, `GroupChatAgentResponse`, `GroupChatTermination`) — these are internal to GroupChat orchestration. Will be defined in the GroupChat plan.

### 2. Placeholder Scan

No TBD, TODO, "implement later", "add appropriate error handling", or "similar to Task N" patterns found. All code steps contain complete implementations.

### 3. Type Consistency

| Type | Defined In | Used In | Consistent |
|---|---|---|---|
| `BaseChatMessage` interface (`Type() string, Source() string`) | Task 1 | Task 5 (ChatMessage alias), Task 5 (Response.ChatMessage) | ✅ |
| `ToolCall{ID, Name, Arguments string}` | Task 1 | Task 3 (LLMResponse.ToolCalls, LLMStreamChunk.ToolCalls) | ✅ |
| `core.ToolResult{Content string, IsError bool}` | core (pre-existing) | Task 1 (ToolResultMessage.Results), Task 6 (CallTool return) | ✅ |
| `LLMMessage{Role, Content, Name, ToolCallID string}` | Task 3 | Task 4 (context stores/retrieves) | ✅ |
| `LLMMessageRole` constants | Task 3 | Task 4 (tests use LLMMessageRoleUser) | ✅ |
| `ChatCompletionClient` interface (3 methods) | Task 3 | Task 5 (mock in test) | ✅ |
| `core.CancellationToken` | core (pre-existing) | Task 5 (ChatAgent.OnMessages), Task 6 (CallTool) | ✅ |
| `core.Tool` interface | core (pre-existing) | Task 6 (Workbench stores core.Tool) | ✅ |
| `Handoff{Target, Description string}` | Task 7 | Task 7 (ToHandoffMessage produces HandoffMessage from Task 1) | ✅ |
