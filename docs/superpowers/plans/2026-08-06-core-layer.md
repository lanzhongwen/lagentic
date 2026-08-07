# Core Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `lagentic-core` — the actor framework and message routing layer with no LLM awareness.

**Architecture:** Two-layer AutoGen-inspired design. Layer 1 (`core`) provides message passing, agent lifecycle, type-based handler dispatch, pub/sub subscriptions, cancellation, and a tool interface. All types use Go interfaces for testability; the single-threaded runtime is the first concrete implementation.

**Tech Stack:** Go 1.24, stdlib only (reflect, sync, context, encoding/json, fmt)

## Global Constraints

- `core` package imports stdlib only — never imports `agentchat` or `ext`
- All errors wrapped with context: `fmt.Errorf("doing X: %w", err)`
- Sentinel errors for distinguishable failure modes, custom error types with `Unwrap()`
- All concurrent code must pass `-race`
- Table-driven tests for multi-scenario validation
- Test names describe behavior: `TestX_Behavior_ExpectedResult`
- Conventional commit format: `feat(scope): description`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `go.mod` | Module definition |
| `core/agent_id.go` | `AgentID` value type |
| `core/topic.go` | `TopicID`, `Subscription` interface, `TypeSubscription` |
| `core/message.go` | `CancellationToken`, `MessageContext` |
| `core/errors.go` | Sentinel errors, `AgentError` |
| `core/agent.go` | `AgentMetadata`, `Agent` interface, `AgentFactory`, `HandlerFunc`, `AgentRuntime` interface, `BaseAgent`, `RoutedAgent` |
| `core/runtime_single.go` | `SingleThreadedAgentRuntime` |
| `core/tool.go` | `ToolSchema`, `ToolResult`, `Tool` interface, `FunctionTool` |

Corresponding `*_test.go` files for each.

---

### Task 1: Go Module + Value Types

**Files:**
- Create: `go.mod`
- Create: `core/agent_id.go`
- Create: `core/agent_id_test.go`
- Create: `core/topic.go`
- Create: `core/topic_test.go`
- Create: `core/message.go`
- Create: `core/message_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `AgentID{Type, Key}`, `TopicID{Type, Source}`, `CancellationToken`, `MessageContext`

- [ ] **Step 1: Initialize Go module**

```bash
cd /home/lzw/projects/src/github.com/lanzhongwen/lagentic
go mod init github.com/lanzhongwen/lagentic
```

- [ ] **Step 2: Write failing tests for AgentID**

Create `core/agent_id_test.go`:

```go
package core

import "testing"

func TestAgentID_String_Format(t *testing.T) {
	id := AgentID{Type: "coordinator", Key: "team-1"}
	want := "coordinator/team-1"
	if got := id.String(); got != want {
		t.Errorf("AgentID.String() = %q, want %q", got, want)
	}
}

func TestAgentID_String_Empty(t *testing.T) {
	id := AgentID{}
	want := "/"
	if got := id.String(); got != want {
		t.Errorf("AgentID.String() = %q, want %q", got, want)
	}
}

func TestAgentID_Equal(t *testing.T) {
	a := AgentID{Type: "coder", Key: "default"}
	b := AgentID{Type: "coder", Key: "default"}
	c := AgentID{Type: "reviewer", Key: "default"}
	if a != b {
		t.Error("identical AgentIDs should be equal")
	}
	if a == c {
		t.Error("different AgentIDs should not be equal")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./core/ -run TestAgentID -v`
Expected: FAIL — `AgentID` undefined

- [ ] **Step 4: Implement AgentID**

Create `core/agent_id.go`:

```go
package core

import "fmt"

// AgentID uniquely identifies an agent instance.
type AgentID struct {
	Type string // agent type (e.g., "coordinator", "coder")
	Key  string // instance key (e.g., team UUID)
}

func (id AgentID) String() string {
	return fmt.Sprintf("%s/%s", id.Type, id.Key)
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./core/ -run TestAgentID -v`
Expected: PASS

- [ ] **Step 6: Write failing tests for TopicID**

Create `core/topic_test.go`:

```go
package core

import "testing"

func TestTopicID_String_Format(t *testing.T) {
	id := TopicID{Type: "task", Source: "coordinator"}
	want := "task/coordinator"
	if got := id.String(); got != want {
		t.Errorf("TopicID.String() = %q, want %q", got, want)
	}
}

func TestTopicID_Equal(t *testing.T) {
	a := TopicID{Type: "task", Source: "coord"}
	b := TopicID{Type: "task", Source: "coord"}
	c := TopicID{Type: "result", Source: "coord"}
	if a != b {
		t.Error("identical TopicIDs should be equal")
	}
	if a == c {
		t.Error("different TopicIDs should not be equal")
	}
}
```

- [ ] **Step 7: Run test to verify it fails**

Run: `go test ./core/ -run TestTopicID -v`
Expected: FAIL — `TopicID` undefined

- [ ] **Step 8: Implement TopicID**

Create `core/topic.go`:

```go
package core

import "fmt"

// TopicID identifies a pub/sub topic.
type TopicID struct {
	Type   string // topic type
	Source string // namespace/context
}

func (id TopicID) String() string {
	return fmt.Sprintf("%s/%s", id.Type, id.Source)
}
```

- [ ] **Step 9: Run test to verify it passes**

Run: `go test ./core/ -run TestTopicID -v`
Expected: PASS

- [ ] **Step 10: Write failing tests for CancellationToken**

Create `core/message_test.go`:

```go
package core

import (
	"testing"
	"time"
)

func TestCancellationToken_Done_BlocksBeforeCancel(t *testing.T) {
	ct := NewCancellationToken()
	select {
	case <-ct.Done():
		t.Error("Done() should block before Cancel()")
	default:
		// expected
	}
}

func TestCancellationToken_Done_ClosesAfterCancel(t *testing.T) {
	ct := NewCancellationToken()
	ct.Cancel()
	select {
	case <-ct.Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() should be closed after Cancel()")
	}
}

func TestCancellationToken_Cancel_Idempotent(t *testing.T) {
	ct := NewCancellationToken()
	ct.Cancel()
	ct.Cancel() // should not panic
	select {
	case <-ct.Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() should be closed after Cancel()")
	}
}

func TestMessageContext_Fields(t *testing.T) {
	ct := NewCancellationToken()
	mc := MessageContext{
		IsRPC:             true,
		TopicID:           TopicID{Type: "task", Source: "coord"},
		Sender:            AgentID{Type: "coordinator", Key: "team-1"},
		CancellationToken: ct,
	}
	if !mc.IsRPC {
		t.Error("IsRPC should be true")
	}
	if mc.TopicID.Type != "task" {
		t.Errorf("TopicID.Type = %q, want %q", mc.TopicID.Type, "task")
	}
	if mc.Sender.Type != "coordinator" {
		t.Errorf("Sender.Type = %q, want %q", mc.Sender.Type, "coordinator")
	}
	if mc.CancellationToken != ct {
		t.Error("CancellationToken should match")
	}
}
```

- [ ] **Step 11: Run test to verify it fails**

Run: `go test ./core/ -run TestCancellationToken -v`
Expected: FAIL — `NewCancellationToken` undefined

- [ ] **Step 12: Implement CancellationToken and MessageContext**

Create `core/message.go`:

```go
package core

import "sync"

// CancellationToken propagates cancellation from user (Ctrl+C) through all layers.
type CancellationToken struct {
	done chan struct{}
	once sync.Once
}

// NewCancellationToken creates an un-cancelled token.
func NewCancellationToken() *CancellationToken {
	return &CancellationToken{
		done: make(chan struct{}),
	}
}

// Cancel closes the done channel. Safe to call multiple times.
func (ct *CancellationToken) Cancel() {
	ct.once.Do(func() { close(ct.done) })
}

// Done returns a channel that is closed when Cancel is called.
func (ct *CancellationToken) Done() <-chan struct{} {
	return ct.done
}

// MessageContext carries per-message metadata.
type MessageContext struct {
	IsRPC             bool
	TopicID           TopicID
	Sender            AgentID
	CancellationToken *CancellationToken
}
```

- [ ] **Step 13: Run all tests to verify they pass**

Run: `go test ./core/ -v`
Expected: All PASS

- [ ] **Step 14: Commit**

```bash
git add go.mod go.sum core/
git commit -m "feat(core): add AgentID, TopicID, CancellationToken, MessageContext"
```

---

### Task 2: Core Errors

**Files:**
- Create: `core/errors.go`
- Create: `core/errors_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `ErrAgentNotFound`, `ErrToolNotFound`, `ErrContextCanceled`, `AgentError`

- [ ] **Step 1: Write failing tests for core errors**

Create `core/errors_test.go`:

```go
package core

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_Values(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		want  string
	}{
		{"AgentNotFound", ErrAgentNotFound, "agent not found"},
		{"ToolNotFound", ErrToolNotFound, "tool not found"},
		{"ContextCanceled", ErrContextCanceled, "context canceled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestAgentError_ErrorFormat(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := &AgentError{
		Agent:  "coder",
		TaskID: "task-42",
		Phase:  "tool_call",
		Cause:  cause,
	}
	want := `agent "coder" task "task-42" phase "tool_call": connection refused`
	if err.Error() != want {
		t.Errorf("AgentError.Error() = %q, want %q", err.Error(), want)
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := &AgentError{
		Agent:  "coder",
		TaskID: "task-42",
		Phase:  "llm_call",
		Cause:  cause,
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is should find the cause through Unwrap")
	}
}

func TestAgentError_WrappedInFmtError(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	agentErr := &AgentError{
		Agent:  "coder",
		TaskID: "task-42",
		Phase:  "llm_call",
		Cause:  cause,
	}
	wrapped := fmt.Errorf("processing failed: %w", agentErr)
	if !errors.Is(wrapped, agentErr) {
		t.Error("errors.Is should find AgentError through fmt.Errorf wrap")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestSentinel -v`
Expected: FAIL — `ErrAgentNotFound` undefined

- [ ] **Step 3: Implement core errors**

Create `core/errors.go`:

```go
package core

import "fmt"

// Sentinel errors for distinguishable failure modes.
var (
	ErrAgentNotFound    = fmt.Errorf("agent not found")
	ErrToolNotFound     = fmt.Errorf("tool not found")
	ErrContextCanceled  = fmt.Errorf("context canceled")
)

// AgentError wraps an error with agent context.
type AgentError struct {
	Agent  string
	TaskID string
	Phase  string // "llm_call", "tool_call", "delegation"
	Cause  error
}

func (e *AgentError) Error() string {
	return fmt.Sprintf("agent %q task %q phase %q: %v", e.Agent, e.TaskID, e.Phase, e.Cause)
}

func (e *AgentError) Unwrap() error { return e.Cause }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run "TestSentinel|TestAgentError" -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add core/errors.go core/errors_test.go
git commit -m "feat(core): add sentinel errors and AgentError type"
```

---

### Task 3: Subscription

**Files:**
- Modify: `core/topic.go` (add Subscription interface, TypeSubscription)
- Modify: `core/topic_test.go` (add tests)

**Interfaces:**
- Consumes: `AgentID`, `TopicID` (from Task 1)
- Produces: `Subscription` interface, `TypeSubscription`

- [ ] **Step 1: Write failing tests for TypeSubscription**

Append to `core/topic_test.go`:

```go
func TestTypeSubscription_IsMatch_MatchingType(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "task", Source: "coordinator"}
	if !sub.IsMatch(topic) {
		t.Error("should match when topic type equals subscription type")
	}
}

func TestTypeSubscription_IsMatch_NonMatchingType(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "result", Source: "coordinator"}
	if sub.IsMatch(topic) {
		t.Error("should not match when topic type differs")
	}
}

func TestTypeSubscription_MapToAgent(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "task", Source: "any"}
	got := sub.MapToAgent(topic)
	want := AgentID{Type: "coordinator", Key: "default"}
	if got != want {
		t.Errorf("MapToAgent() = %v, want %v", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestTypeSubscription -v`
Expected: FAIL — `NewTypeSubscription` undefined

- [ ] **Step 3: Implement Subscription interface and TypeSubscription**

Append to `core/topic.go`:

```go
// Subscription maps topics to agents.
type Subscription interface {
	IsMatch(topic TopicID) bool
	MapToAgent(topic TopicID) AgentID
}

// TypeSubscription matches topics by Type and maps them to a specific agent type.
type TypeSubscription struct {
	topicType  string
	agentType  string
	agentKey   string
}

// NewTypeSubscription creates a subscription that matches topics with the given type
// and maps them to an agent of the given type and key.
func NewTypeSubscription(topicType, agentType string, agentKey ...string) *TypeSubscription {
	key := "default"
	if len(agentKey) > 0 {
		key = agentKey[0]
	}
	return &TypeSubscription{
		topicType: topicType,
		agentType: agentType,
		agentKey:  key,
	}
}

func (s *TypeSubscription) IsMatch(topic TopicID) bool {
	return topic.Type == s.topicType
}

func (s *TypeSubscription) MapToAgent(topic TopicID) AgentID {
	return AgentID{Type: s.agentType, Key: s.agentKey}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run TestTypeSubscription -v`
Expected: All PASS

- [ ] **Step 5: Run all core tests with race detector**

Run: `go test -race ./core/ -v`
Expected: All PASS, no races

- [ ] **Step 6: Commit**

```bash
git add core/topic.go core/topic_test.go
git commit -m "feat(core): add Subscription interface and TypeSubscription"
```

---

### Task 4: Agent Interface + BaseAgent

**Files:**
- Create: `core/agent.go`
- Create: `core/agent_test.go`

**Interfaces:**
- Consumes: `AgentID`, `MessageContext` (from Task 1), `CancellationToken` (from Task 1)
- Produces: `AgentMetadata`, `Agent` interface, `AgentFactory`, `HandlerFunc`, `AgentRuntime` interface, `BaseAgent`

- [ ] **Step 1: Write failing tests for BaseAgent**

Create `core/agent_test.go`:

```go
package core

import (
	"context"
	"testing"
)

// mockRuntime is a test double for AgentRuntime.
type mockRuntime struct {
	sentMessages   []sentMessage
	publishedMsgs  []publishedMessage
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestBaseAgent -v`
Expected: FAIL — `AgentMetadata`, `Agent`, `BaseAgent` undefined

- [ ] **Step 3: Implement Agent interface, AgentRuntime interface, and BaseAgent**

Create `core/agent.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run TestBaseAgent -v`
Expected: All PASS

- [ ] **Step 5: Run all core tests with race detector**

Run: `go test -race ./core/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add core/agent.go core/agent_test.go
git commit -m "feat(core): add Agent interface, AgentRuntime interface, BaseAgent"
```

---

### Task 5: RoutedAgent

**Files:**
- Modify: `core/agent.go` (add RoutedAgent, handlerEntry)
- Modify: `core/agent_test.go` (add RoutedAgent tests)

**Interfaces:**
- Consumes: `BaseAgent`, `HandlerFunc`, `MessageContext` (from Task 4)
- Produces: `RoutedAgent` (with `RegisterRPCHandler`, `RegisterEventHandler`)

- [ ] **Step 1: Write failing tests for RoutedAgent**

Append to `core/agent_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestRoutedAgent -v`
Expected: FAIL — `NewRoutedAgent` undefined

- [ ] **Step 3: Implement RoutedAgent**

Append to `core/agent.go`:

```go
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
		return nil, fmt.Errorf("no handler registered for message type %v", msgType)
	}

	var lastResult any
	for _, entry := range entries {
		if mc.IsRPC && !entry.isRPC {
			continue // skip event handlers for RPC messages
		}
		if !mc.IsRPC && entry.isRPC {
			continue // skip RPC handlers for event messages
		}
		result, err := entry.handler(ctx, msg, mc)
		if err != nil {
			return nil, fmt.Errorf("handler for %v: %w", msgType, err)
		}
		lastResult = result
	}
	return lastResult, nil
}
```

Add `"reflect"` to the import block in `core/agent.go`. The full import block becomes:

```go
import (
	"context"
	"fmt"
	"reflect"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run TestRoutedAgent -v`
Expected: All PASS

- [ ] **Step 5: Run all core tests with race detector**

Run: `go test -race ./core/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add core/agent.go core/agent_test.go
git commit -m "feat(core): add RoutedAgent with type-based handler dispatch"
```

---

### Task 6: SingleThreadedAgentRuntime

**Files:**
- Create: `core/runtime_single.go`
- Create: `core/runtime_single_test.go`

**Interfaces:**
- Consumes: `Agent`, `AgentFactory`, `AgentRuntime`, `AgentID`, `TopicID`, `Subscription`, `MessageContext`, `ErrAgentNotFound` (from Tasks 1-5)
- Produces: `SingleThreadedAgentRuntime` — the concrete runtime implementation

- [ ] **Step 1: Write failing tests for SingleThreadedAgentRuntime**

Create `core/runtime_single_test.go`:

```go
package core

import (
	"context"
	"errors"
	"testing"
	"time"
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestSingleThreadedAgentRuntime -v`
Expected: FAIL — `NewSingleThreadedAgentRuntime` undefined

- [ ] **Step 3: Implement SingleThreadedAgentRuntime**

Create `core/runtime_single.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run TestSingleThreadedAgentRuntime -v`
Expected: All PASS

- [ ] **Step 5: Run all core tests with race detector**

Run: `go test -race ./core/ -v`
Expected: All PASS, no races

- [ ] **Step 6: Commit**

```bash
git add core/runtime_single.go core/runtime_single_test.go
git commit -m "feat(core): add SingleThreadedAgentRuntime with RPC and pub/sub"
```

---

### Task 7: Tool System

**Files:**
- Create: `core/tool.go`
- Create: `core/tool_test.go`

**Interfaces:**
- Consumes: `CancellationToken` (from Task 1)
- Produces: `ToolSchema`, `ToolResult`, `Tool` interface, `FunctionTool`

- [ ] **Step 1: Write failing tests for FunctionTool**

Create `core/tool_test.go`:

```go
package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestFunctionTool_Name(t *testing.T) {
	tool := NewFunctionTool("read_file", "Read a file", ToolSchema{
		Name:        "read_file",
		Description: "Read a file",
		Parameters:  map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *CancellationToken) (any, error) {
		return nil, nil
	})
	if tool.Name() != "read_file" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "read_file")
	}
}

func TestFunctionTool_Description(t *testing.T) {
	tool := NewFunctionTool("read_file", "Read a file", ToolSchema{
		Name:        "read_file",
		Description: "Read a file",
		Parameters:  map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *CancellationToken) (any, error) {
		return nil, nil
	})
	if tool.Description() != "Read a file" {
		t.Errorf("Description() = %q, want %q", tool.Description(), "Read a file")
	}
}

func TestFunctionTool_Schema(t *testing.T) {
	schema := ToolSchema{
		Name:        "read_file",
		Description: "Read a file",
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{"path": map[string]any{"type": "string"}}},
	}
	tool := NewFunctionTool("read_file", "Read a file", schema, func(_ context.Context, _ json.RawMessage, _ *CancellationToken) (any, error) {
		return nil, nil
	})
	got := tool.Schema()
	if got.Name != schema.Name {
		t.Errorf("Schema().Name = %q, want %q", got.Name, schema.Name)
	}
	if got.Description != schema.Description {
		t.Errorf("Schema().Description = %q, want %q", got.Description, schema.Description)
	}
}

func TestFunctionTool_RunJSON_Executes(t *testing.T) {
	var receivedArgs json.RawMessage
	tool := NewFunctionTool("echo", "Echo input", ToolSchema{
		Name:        "echo",
		Description: "Echo input",
		Parameters:  map[string]any{"type": "object"},
	}, func(_ context.Context, args json.RawMessage, _ *CancellationToken) (any, error) {
		receivedArgs = args
		return map[string]any{"echo": true}, nil
	})

	args := json.RawMessage(`{"msg": "hello"}`)
	result, err := tool.RunJSON(context.Background(), args, nil)
	if err != nil {
		t.Fatalf("RunJSON() error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T, want map[string]any", result)
	}
	if !m["echo"].(bool) {
		t.Error("result[\"echo\"] = false, want true")
	}
	if string(receivedArgs) != `{"msg": "hello"}` {
		t.Errorf("received args = %s, want %s", receivedArgs, `{"msg": "hello"}`)
	}
}

func TestFunctionTool_RunJSON_PropagatesError(t *testing.T) {
	toolErr := errors.New("tool failed")
	tool := NewFunctionTool("fail", "Always fails", ToolSchema{
		Name:        "fail",
		Description: "Always fails",
		Parameters:  map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, _ *CancellationToken) (any, error) {
		return nil, toolErr
	})

	_, err := tool.RunJSON(context.Background(), json.RawMessage(`{}`), nil)
	if !errors.Is(err, toolErr) {
		t.Errorf("error = %v, want wrapped %v", err, toolErr)
	}
}

func TestFunctionTool_RunJSON_RespectsCancellation(t *testing.T) {
	ct := NewCancellationToken()
	ct.Cancel()

	tool := NewFunctionTool("slow", "Slow tool", ToolSchema{
		Name:        "slow",
		Description: "Slow tool",
		Parameters:  map[string]any{"type": "object"},
	}, func(_ context.Context, _ json.RawMessage, cancel *CancellationToken) (any, error) {
		select {
		case <-cancel.Done():
			return nil, ErrContextCanceled
		default:
			return "done", nil
		}
	})

	_, err := tool.RunJSON(context.Background(), json.RawMessage(`{}`), ct)
	if !errors.Is(err, ErrContextCanceled) {
		t.Errorf("error = %v, want ErrContextCanceled", err)
	}
}

func TestToolResult(t *testing.T) {
	tr := ToolResult{Content: "file contents here", IsError: false}
	if tr.Content != "file contents here" {
		t.Errorf("Content = %q, want %q", tr.Content, "file contents here")
	}
	if tr.IsError {
		t.Error("IsError should be false")
	}

	trErr := ToolResult{Content: "file not found", IsError: true}
	if !trErr.IsError {
		t.Error("IsError should be true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/ -run TestFunctionTool -v`
Expected: FAIL — `ToolSchema`, `FunctionTool` undefined

- [ ] **Step 3: Implement Tool system**

Create `core/tool.go`:

```go
package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolSchema describes a tool's name, purpose, and input parameters.
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema object
}

// ToolResult is the output of a tool execution.
type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error,omitempty"`
}

// Tool is the base tool interface.
type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	RunJSON(ctx context.Context, args json.RawMessage, ct *CancellationToken) (any, error)
}

// toolFunc is the function signature for FunctionTool callbacks.
type toolFunc func(ctx context.Context, args json.RawMessage, ct *CancellationToken) (any, error)

// FunctionTool wraps a Go function into a Tool.
type FunctionTool struct {
	name        string
	description string
	schema      ToolSchema
	fn          toolFunc
}

// NewFunctionTool creates a tool from a name, description, schema, and function.
func NewFunctionTool(name, description string, schema ToolSchema, fn toolFunc) *FunctionTool {
	return &FunctionTool{
		name:        name,
		description: description,
		schema:      schema,
		fn:          fn,
	}
}

func (t *FunctionTool) Name() string        { return t.name }
func (t *FunctionTool) Description() string  { return t.description }
func (t *FunctionTool) Schema() ToolSchema   { return t.schema }

func (t *FunctionTool) RunJSON(ctx context.Context, args json.RawMessage, ct *CancellationToken) (any, error) {
	if ct != nil {
		select {
		case <-ct.Done():
			return nil, fmt.Errorf("tool %q: %w", t.name, ErrContextCanceled)
		default:
		}
	}
	result, err := t.fn(ctx, args, ct)
	if err != nil {
		return nil, fmt.Errorf("tool %q: %w", t.name, err)
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./core/ -run "TestFunctionTool|TestToolResult" -v`
Expected: All PASS

- [ ] **Step 5: Run all core tests with race detector**

Run: `go test -race ./core/ -v`
Expected: All PASS, no races

- [ ] **Step 6: Run the full test suite**

Run: `go test ./... -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add core/tool.go core/tool_test.go
git commit -m "feat(core): add Tool interface, ToolSchema, ToolResult, FunctionTool"
```

---

## Self-Review

### 1. Spec Coverage

| Spec Section | Covered By Task |
|---|---|
| `AgentID` struct | Task 1 |
| `TopicID` struct | Task 1 |
| `MessageContext` struct | Task 1 |
| `CancellationToken` | Task 1 |
| `Agent` interface | Task 4 |
| `BaseAgent` with SendMessage/PublishMessage | Task 4 |
| `RoutedAgent` with RegisterRPCHandler/RegisterEventHandler | Task 5 |
| `AgentRuntime` interface | Task 4 (interface), Task 6 (impl) |
| `SingleThreadedAgentRuntime` | Task 6 |
| `Subscription` interface | Task 3 |
| `TypeSubscription` | Task 3 |
| `AgentMetadata` | Task 4 |
| `AgentFactory` | Task 4 |
| `HandlerFunc` | Task 4 |
| `Tool` interface | Task 7 |
| `FunctionTool` | Task 7 |
| `ToolSchema` | Task 7 |
| `ErrAgentNotFound` | Task 2 |
| `ErrToolNotFound` | Task 2 |
| `ErrContextCanceled` | Task 2 |
| `AgentError` | Task 2 |
| Cancellation flow (runtime checks `ct.Done()`) | Task 6 |

**Gaps:**
- `ErrProviderNotFound`, `ErrMaxTurnsExceeded`, `ErrTokenLimitExceeded` — these belong to `agentchat`/`ext` layers, not `core`. Will be added in subsequent plans.
- `ProviderError` — belongs in `ext` layer.
- `Workbench` — spec places it in `agentchat`, not `core`. Will be added in AgentChat plan.
- `SaveState`/`LoadState` on agents — interface defined, trivial default implementations provided. Full state serialization deferred to when agents have meaningful state.

### 2. Placeholder Scan

No TBD, TODO, "implement later", "add appropriate error handling", or "similar to Task N" patterns found. All code steps contain complete implementations.

### 3. Type Consistency

| Type | Defined In | Used In | Consistent |
|---|---|---|---|
| `AgentID{Type, Key}` | Task 1 | Tasks 2-7 | ✅ |
| `TopicID{Type, Source}` | Task 1 | Tasks 3, 4, 6 | ✅ |
| `MessageContext{IsRPC, TopicID, Sender, CancellationToken}` | Task 1 | Tasks 4, 5, 6 | ✅ |
| `CancellationToken` with `Cancel()`, `Done()` | Task 1 | Tasks 6, 7 | ✅ |
| `HandlerFunc` signature | Task 4 | Task 5 | ✅ |
| `AgentRuntime` interface | Task 4 | Tasks 4, 5, 6 | ✅ |
| `ToolSchema{Name, Description, Parameters}` | Task 7 | Task 7 | ✅ |
| `FunctionTool.RunJSON(ctx, args, ct)` | Task 7 | Task 7 | ✅ |
