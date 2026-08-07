# Lagentic Design Specification

**Date:** 2026-08-06
**Status:** Approved
**Scope:** Framework-first prototype with 4 agents (Coordinator, Coder, Reviewer, Tester)

## Overview

Lagentic is a CLI-based multi-agent coding system. Each agent has an independent context with a different model. The system coordinates specialized AI agents to perform software development tasks end-to-end.

The architecture follows AutoGen's two-layer design pattern, adapted for Go.

## Architecture: Two-Layer Design

### Layer 1: `lagentic-core` — Actor Framework & Message Routing

The runtime layer. No LLM awareness — just message passing, agent lifecycle, and subscriptions.

```go
// AgentID uniquely identifies an agent instance.
type AgentID struct {
    Type string // agent type (e.g., "coordinator", "coder")
    Key  string // instance key (e.g., team UUID)
}

// TopicID identifies a pub/sub topic.
type TopicID struct {
    Type   string // topic type
    Source string // namespace/context
}

// MessageContext carries per-message metadata.
type MessageContext struct {
    IsRPC             bool
    TopicID           TopicID
    Sender            AgentID
    CancellationToken *CancellationToken
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

// BaseAgent adds runtime binding and send/publish capabilities.
// SendMessage and PublishMessage fill in the sender AgentID automatically from a.id.
type BaseAgent struct {
    id      AgentID
    runtime AgentRuntime
}
func (a *BaseAgent) SendMessage(ctx context.Context, msg any, recipient AgentID) (any, error)
func (a *BaseAgent) PublishMessage(ctx context.Context, msg any, topic TopicID) error

// RoutedAgent adds type-based message routing.
// Handlers registered via RegisterRPCHandler / RegisterEventHandler instead of decorators.
type RoutedAgent struct {
    BaseAgent
    handlers map[reflect.Type][]handlerEntry
}
func (a *RoutedAgent) OnMessage(ctx context.Context, msg any, mc MessageContext) (any, error)
func (a *RoutedAgent) RegisterRPCHandler(msgType reflect.Type, handler HandlerFunc)
func (a *RoutedAgent) RegisterEventHandler(msgType reflect.Type, handler HandlerFunc)

// AgentRuntime is the central orchestrator protocol.
type AgentRuntime interface {
    SendMessage(ctx context.Context, msg any, recipient AgentID, sender AgentID) (any, error)
    PublishMessage(ctx context.Context, msg any, topic TopicID, sender AgentID) error
    RegisterFactory(agentType string, factory AgentFactory, subscriptions ...Subscription) error
    AddSubscription(sub Subscription) error
}

// Subscription maps topics to agents.
type Subscription interface {
    IsMatch(topic TopicID) bool
    MapToAgent(topic TopicID) AgentID
}
```

### Layer 2: `lagentic-agentchat` — Chat Agents, Teams, and LLM Integration

Built on top of core. Adds LLM awareness, tools, conversation history.

```go
// ChatAgent is the high-level agent interface.
type ChatAgent interface {
    Name() string
    Description() string
    OnMessages(ctx context.Context, messages []ChatMessage, ct *CancellationToken) (Response, error)
    OnMessagesStream(ctx context.Context, messages []ChatMessage, ct *CancellationToken) (<-chan AgentEvent, error)
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

// AssistantAgent wraps an LLM client with tools and handoffs.
type AssistantAgent struct {
    name        string
    model       ChatCompletionClient
    context     ChatCompletionContext
    workbench   Workbench
    handoffs    []Handoff
    systemMsg   string
    maxToolIter int
    // Memory is deferred to a later iteration. Context enrichment is via ChatCompletionContext only.
}

// ChatCompletionClient abstracts LLM providers.
type ChatCompletionClient interface {
    Create(ctx context.Context, messages []LLMMessage, options ...CompletionOption) (LLMResponse, error)
    CreateStream(ctx context.Context, messages []LLMMessage, options ...CompletionOption) (<-chan LLMStreamChunk, error)
    ModelInfo() ModelInfo
}

// ChatCompletionContext manages per-agent conversation history.
type ChatCompletionContext interface {
    AddMessage(msg LLMMessage) error
    GetMessages() []LLMMessage
    Clear() error
    SaveState() (map[string]any, error)
    LoadState(state map[string]any) error
}
// Implementations: Unbounded, Buffered(N), TokenLimited
```

### Tool System

```go
// Tool is the base tool interface.
type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    RunJSON(ctx context.Context, args json.RawMessage, ct *CancellationToken) (any, error)
}

// Workbench is a container for tools.
type Workbench interface {
    ListTools() []ToolSchema
    CallTool(ctx context.Context, name string, args json.RawMessage, ct *CancellationToken) (ToolResult, error)
    Register(tool Tool) error
}

// FunctionTool wraps a Go function into a Tool.
type FunctionTool struct { /* ... */ }
```

### Team & GroupChat

```go
// Team is a group of agents with orchestration.
type Team interface {
    Name() string
    Run(ctx context.Context, task string, ct *CancellationToken) (TaskResult, error)
    RunStream(ctx context.Context, task string, ct *CancellationToken) (<-chan AgentEvent, error)
    Reset(ctx context.Context) error
    SaveState() (map[string]any, error)
    LoadState(state map[string]any) error
}

// BaseGroupChat is the foundation for all team types.
type BaseGroupChat struct {
    participants []ChatAgent
    runtime      AgentRuntime
    manager      GroupChatManager
    termination  TerminationCondition
}

// GroupChatManager handles speaker selection and message thread.
type GroupChatManager interface {
    SelectSpeaker(thread []ChatMessage) ([]string, error)
    Reset() error
}

// TerminationCondition decides when the team stops. Composable via And()/Or().
type TerminationCondition interface {
    Check(messages []ChatMessage) (*StopMessage, error)
    Reset() error
}

// Concrete team types:
type RoundRobinGroupChat struct { BaseGroupChat }
type SelectorGroupChat struct {
    BaseGroupChat
    selectorModel  ChatCompletionClient
    selectorPrompt string
}
```

### Message Types

```go
// BaseChatMessage — all agent-to-agent messages.
type BaseChatMessage interface {
    Type() string
    Source() string
}

type TextMessage struct { Content string; Source string }
type HandoffMessage struct { Target string; Context string; Source string }
type ToolCallMessage struct { ToolCalls []ToolCall; Source string }
type ToolResultMessage struct { Results []ToolResult; Source string }
type StopMessage struct { Content string; Source string }

// Internal GroupChat events (not user-facing):
type GroupChatStart struct { Messages []BaseChatMessage }
type GroupChatRequestPublish struct {}
type GroupChatAgentResponse struct { Response Response }
type GroupChatTermination struct { Reason string }
```

## The 4-Agent Prototype

### Agent Definitions

```go
coordinator := NewAssistantAgent(
    "coordinator",
    WithModel(anthropicClient, "claude-sonnet-4-6"),
    WithSystemPrompt("You are the coordinator. Break down tasks and delegate to the right agent..."),
    WithWorkbench(toolWorkbench), // file listing, code search tools only
)

coder := NewAssistantAgent(
    "coder",
    WithModel(anthropicClient, "claude-sonnet-4-6"),
    WithSystemPrompt("You are a software developer. Write clean, correct code..."),
    WithWorkbench(coderWorkbench), // file read/write, shell exec, search
)

reviewer := NewAssistantAgent(
    "reviewer",
    WithModel(anthropicClient, "claude-sonnet-4-6"),
    WithSystemPrompt("You are a code reviewer. Identify bugs, security issues, style problems..."),
    WithWorkbench(readOnlyWorkbench), // file read + search only — no write access
)

tester := NewAssistantAgent(
    "tester",
    WithModel(anthropicClient, "claude-sonnet-4-6"),
    WithSystemPrompt("You are a test engineer. Write tests, run them, report results..."),
    WithWorkbench(coderWorkbench), // file read/write (for test files), shell exec (to run tests)
)
```

### Tool Workbench Per Agent

| Workbench | Tools | Agents |
|-----------|-------|--------|
| `toolWorkbench` | `list_files`, `search_code` | Coordinator |
| `coderWorkbench` | `read_file`, `write_file`, `run_shell`, `search_code` | Coder, Tester |
| `readOnlyWorkbench` | `read_file`, `search_code` | Reviewer |

### Orchestration: SelectorGroupChat

The coordinator (as LLM-based selector) picks the next speaker based on conversation context. This naturally handles feedback loops:

```
User: "Implement a Go HTTP user service"
  │
  ├─ Coordinator: "Coder, implement the HTTP service. Spec: ..."
  │     ▼
  ├─ Coder: writes main.go, handlers.go → ToolCall: write_file
  │     ▼
  ├─ Reviewer: "Found issues: 1) Missing error handling, 2) SQL injection risk"
  │     ▼
  ├─ Coordinator: "Coder, fix the issues found by Reviewer"
  │     ▼
  ├─ Coder: fixes code → ToolCall: write_file
  │     ▼
  ├─ Reviewer: "All issues resolved. Code looks good."
  │     ▼
  ├─ Coordinator: "Tester, write and run tests for the HTTP service"
  │     ▼
  ├─ Tester: writes user_test.go → ToolCall: write_file, run_shell("go test")
  │     ▼
  ├─ Tester: "2 tests failed: TestGetUser returns 500, TestCreateUser missing validation"
  │     ▼
  ├─ Coordinator: "Coder, fix the failing tests"
  │     ▼
  ├─ Coder: fixes code → ToolCall: write_file
  │     ▼
  ├─ Coordinator: "Tester, re-run the tests"
  │     ▼
  ├─ Tester: runs tests → "All tests passing"
  │     ▼
  └─ Coordinator: "Task complete. Summary: ..."
```

### Termination Conditions

```go
termination := MaxTurnTermination(30).Or(
    TextMentionTermination("TASK COMPLETE"))
```

### Handoff Support

Agents can explicitly hand off to another agent instead of waiting for the selector:

```go
coder := NewAssistantAgent("coder",
    WithHandoffs(
        Handoff{Target: "reviewer", Description: "Code is ready for review"},
        Handoff{Target: "tester", Description: "Code is ready for testing"},
    ),
)
```

When the Coder produces a `HandoffMessage{Target: "reviewer"}`, the GroupChatManager routes directly to the Reviewer.

### CLI Entry Point

```go
func main() {
    // 1. Create provider clients
    anthropicClient := ext.NewAnthropicClient(apiKey)

    // 2. Create tool workbenches
    coderWB := NewStaticWorkbench(readFile, writeFile, runShell, searchCode)
    reviewWB := NewStaticWorkbench(readFile, searchCode)
    coordWB := NewStaticWorkbench(listFiles, searchCode)

    // 3. Create agents
    coordinator := NewAssistantAgent("coordinator", ...)
    coder := NewAssistantAgent("coder", ...)
    reviewer := NewAssistantAgent("reviewer", ...)
    tester := NewAssistantAgent("tester", ...)

    // 4. Create team
    team := NewSelectorGroupChat(
        []ChatAgent{coordinator, coder, reviewer, tester},
        selectorModel, selectorPrompt,
        WithTermination(termination),
    )

    // 5. Run
    result, err := team.Run(ctx, userRequest, nil)
}
```

## Multi-Provider Architecture

### Extensible Provider Plugin System

```go
// Provider is the plugin interface for adding new LLM providers.
type Provider interface {
    Name() string
    CreateClient(config ProviderConfig) (ChatCompletionClient, error)
}

// ProviderConfig is a string map — each provider defines its own keys.
// GLM: {"api_key_env": "GLM_API_KEY", "base_url": "https://open.bigmodel.cn/api/paas/v4"}
// DeepSeek: {"api_key_env": "DEEPSEEK_API_KEY"}
// Ollama: {"base_url": "http://localhost:11434"}
type ProviderConfig map[string]string

// GlobalProviderRegistry — providers self-register via init().
var GlobalProviderRegistry = NewProviderRegistry()

type ProviderRegistry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *ProviderRegistry) Register(p Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers[p.Name()] = p
}

func (r *ProviderRegistry) CreateClient(providerName string, config ProviderConfig) (ChatCompletionClient, error) {
    r.mu.RLock()
    p, ok := r.providers[providerName]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("unknown provider %q — available: %v", providerName, r.RegisteredNames())
    }
    return p.CreateClient(config)
}
```

### Built-in Providers

```
lagentic-ext/
├── anthropic/        # Claude models — full ChatCompletionClient implementation
├── openai/           # GPT models + OpenAI-compatible adapter (NewCompatibleClient)
├── glm/              # GLM models — delegates to openai.NewCompatibleClient with GLM base URL
├── deepseek/         # DeepSeek models — delegates to openai.NewCompatibleClient with DeepSeek base URL
└── ollama/           # Local models — full ChatCompletionClient implementation
```

### OpenAI-Compatible Shortcut

Many providers (GLM, DeepSeek, Moonshot, Qwen) expose OpenAI-compatible APIs. The `openai` package provides `NewCompatibleClient(config)` that accepts a custom `base_url`, so dedicated provider packages only need to set defaults:

```go
// lagentic-ext/deepseek/provider.go
type DeepSeekProvider struct{}

func (p *DeepSeekProvider) Name() string { return "deepseek" }

func (p *DeepSeekProvider) CreateClient(config ProviderConfig) (ChatCompletionClient, error) {
    config["base_url"] = "https://api.deepseek.com/v1"
    return openai.NewCompatibleClient(config)
}

func init() {
    GlobalProviderRegistry.Register(&DeepSeekProvider{})
}
```

### Adding a New Provider

Three steps, zero framework changes:

1. Implement the `Provider` interface (~50-100 lines for OpenAI-compatible providers)
2. Call `GlobalProviderRegistry.Register()` in `init()`
3. Add the provider config to `lagentic.yaml`

For non-OpenAI-compatible providers, implement `ChatCompletionClient` directly.

### Configuration

```yaml
# lagentic.yaml
providers:
  anthropic:
    api_key_env: ANTHROPIC_API_KEY
  glm:
    api_key_env: GLM_API_KEY
  deepseek:
    api_key_env: DEEPSEEK_API_KEY
  ollama:
    base_url: http://localhost:11434

agents:
  coordinator:
    model: anthropic:claude-sonnet-4-6
  coder:
    model: deepseek:deepseek-coder
  reviewer:
    model: glm:glm-4-plus
  tester:
    model: deepseek:deepseek-chat
```

## Project Structure

```
lagentic/
├── cmd/
│   └── lagentic/           # CLI entry point
│       └── main.go
├── config/
│   └── config.go           # lagentic.yaml loading, validation
├── core/                   # Layer 1: actor framework & message routing
│   ├── agent.go            # Agent, BaseAgent, RoutedAgent
│   ├── runtime.go          # AgentRuntime interface
│   ├── runtime_single.go   # SingleThreadedAgentRuntime
│   ├── message.go          # MessageContext, CancellationToken
│   ├── topic.go            # TopicID, Subscription, TypeSubscription
│   ├── agent_id.go         # AgentID
│   └── tool.go             # Tool interface, FunctionTool
├── agentchat/              # Layer 2: chat agents, teams, LLM integration
│   ├── chat_agent.go       # ChatAgent, Response
│   ├── assistant_agent.go  # AssistantAgent
│   ├── messages.go         # TextMessage, HandoffMessage, StopMessage, etc.
│   ├── model_client.go     # ChatCompletionClient interface
│   ├── model_context.go    # ChatCompletionContext + implementations
│   ├── workbench.go        # Workbench interface, StaticWorkbench
│   ├── handoff.go          # Handoff definition
│   ├── team.go             # Team interface
│   ├── group_chat.go       # BaseGroupChat, GroupChatManager
│   ├── group_chat_events.go
│   ├── selector_group_chat.go
│   ├── round_robin_group_chat.go
│   └── termination.go      # TerminationCondition + And/Or
├── ext/                    # Provider & tool extensions
│   ├── registry.go         # ProviderRegistry, Provider interface
│   ├── anthropic/
│   │   └── provider.go
│   ├── openai/
│   │   └── provider.go
│   ├── glm/
│   │   └── provider.go
│   ├── deepseek/
│   │   └── provider.go
│   ├── ollama/
│   │   └── provider.go
│   └── tools/
│       ├── file.go         # readFile, writeFile, listFiles
│       ├── shell.go        # runShell
│       └── search.go       # searchCode
├── lagentic.yaml
├── go.mod
├── go.sum
└── README.md
```

### Package Dependency Rule

| Package | Depends On | Responsibility |
|---------|-----------|----------------|
| `core` | stdlib only | Message routing, agent lifecycle, subscriptions |
| `agentchat` | `core` | Chat agents, LLM integration, teams, tools |
| `ext/anthropic` | `agentchat` | Anthropic API client |
| `ext/openai` | `agentchat` | OpenAI / OpenAI-compatible API client |
| `ext/glm` | `ext/openai` | GLM defaults + OpenAI-compatible client |
| `ext/deepseek` | `ext/openai` | DeepSeek defaults + OpenAI-compatible client |
| `ext/ollama` | `agentchat` | Ollama API client |
| `ext/tools` | `core` | Built-in tool implementations |
| `config` | `ext`, `agentchat` | Config loading → provider + agent construction |
| `cmd/lagentic` | `config`, `agentchat` | CLI entry point |

**`core` never imports `agentchat` or `ext`. `agentchat` never imports `ext`.** This keeps the core framework provider-agnostic.

### CLI Interface

```
$ lagentic                              # interactive mode (REPL)
$ lagentic "implement a Go HTTP user service"  # one-shot mode
$ lagentic --config ./my-config.yaml    # custom config path
$ lagentic --model deepseek:deepseek-coder  # override default model
$ lagentic --verbose                    # show agent reasoning + tool calls
```

## Error Handling

All errors wrapped with context. Custom error types for distinguishable failure modes.

```go
var (
    ErrAgentNotFound      = errors.New("agent not found")
    ErrProviderNotFound   = errors.New("provider not found")
    ErrToolNotFound       = errors.New("tool not found")
    ErrMaxTurnsExceeded   = errors.New("max turns exceeded")
    ErrContextCanceled    = errors.New("context canceled")
    ErrTokenLimitExceeded = errors.New("token limit exceeded")
)

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

type ProviderError struct {
    Provider string
    Model    string
    Cause    error
}

func (e *ProviderError) Error() string {
    return fmt.Sprintf("provider %q model %q: %v", e.Provider, e.Model, e.Cause)
}
func (e *ProviderError) Unwrap() error { return e.Cause }
```

**Error propagation:**

- Tool execution fails → `AgentError{Phase: "tool_call"}` → agent can retry or report
- LLM call fails → `ProviderError` → `AgentError{Phase: "llm_call"}` → coordinator decides: retry, switch provider, or report to user
- Agent panics → `BaseAgent.OnMessage` recovers, returns error → coordinator handles
- Team exceeds max turns → `MaxTurnTermination` → `ErrMaxTurnsExceeded` returned to CLI

## Cancellation

```go
type CancellationToken struct {
    mu   sync.RWMutex
    done chan struct{}
    once sync.Once
}

func NewCancellationToken() *CancellationToken {
    return &CancellationToken{done: make(chan struct{})}
}
func (ct *CancellationToken) Cancel()   { ct.once.Do(func() { close(ct.done) }) }
func (ct *CancellationToken) Done() <-chan struct{} { return ct.done }
```

**Cancellation flows:**

1. User hits Ctrl+C → signal handler calls `ct.Cancel()`
2. `CancellationToken` passed through `MessageContext` to every agent
3. Long-running tool calls check `ct.Done()` via `select` with `ctx.Done()`
4. LLM streaming calls select on both `ctx.Done()` and stream channel
5. GroupChatManager checks cancellation before each speaker selection

## Observability

```go
type AgentEvent struct {
    Type      string    // "llm_call", "tool_call", "tool_result", "delegation", "speaker_selected"
    Agent     string
    Timestamp time.Time
    Data      any
}

type EventEmitter struct {
    ch chan AgentEvent
}

func (e *EventEmitter) Emit(event AgentEvent) {
    select {
    case e.ch <- event:
    default: // non-blocking drop if consumer is slow
    }
}
```

**Event types:**

| Event Type | Data | Use Case |
|-----------|------|----------|
| `speaker_selected` | `{Speaker, Reason}` | CLI shows which agent is acting |
| `llm_call` | `{Model, InputTokens}` | Token usage tracking |
| `llm_response` | `{OutputTokens, FinishReason}` | Token usage tracking |
| `tool_call` | `{Tool, Args}` | `--verbose` display |
| `tool_result` | `{Tool, Result}` | `--verbose` display |
| `delegation` | `{From, To, Task}` | Trace agent interactions |
| `termination` | `{Reason}` | Final summary |

**CLI display modes:**

- **Normal**: `[agent] summary text` per agent turn
- **--verbose**: tool calls, LLM token counts, delegation details
- **--json**: structured `AgentEvent` JSON per line (for piping/automation)

## Testing Strategy

### By Layer

- **`core`**: Pure message routing tests — no LLM calls. Test RoutedAgent dispatch, pub/sub delivery, cancellation.
- **`agentchat`**: Mock the LLM client via `MockChatCompletionClient` with predetermined responses. Test tool call loops, speaker selection, feedback loops, termination conditions.
- **`ext/*` providers**: Unit tests with `httptest.Server` to mock APIs. Integration tests gated behind `LAGENTIC_INTEGRATION=1` env var.
- **`ext/tools`**: Test against real filesystem with `t.TempDir()`. Shell tool with allowed-command whitelist.

### Test Commands

```bash
go test ./...                    # all unit tests
go test ./core/...               # single package
go test -race ./...              # race detector
LAGENTIC_INTEGRATION=1 go test ./ext/...  # integration tests
go test ./agentchat/ -run TestSelectorGroupChat_FeedbackLoop  # single test
```
