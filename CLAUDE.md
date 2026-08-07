# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lagentic is a **coding-oriented multi-agent collaboration system** implemented as a Go CLI tool (similar to Claude Code / OpenCode). Each agent has an independent context with a different model. The system coordinates specialized AI agents to perform software development tasks end-to-end.

**Current status:** Pre-implementation — approved design spec at `docs/superpowers/specs/2026-08-06-lagentic-design.md`. Language: Go.

## Architecture: Two-Layer Design (AutoGen-inspired)

### Layer 1: `lagentic-core` — Actor Framework & Message Routing
No LLM awareness. Pure message passing, agent lifecycle, and subscriptions.

Key types: `AgentID`, `TopicID`, `MessageContext`, `Agent` interface, `BaseAgent` (send/publish), `RoutedAgent` (type-based handler dispatch via `RegisterRPCHandler`/`RegisterEventHandler`), `AgentRuntime` interface, `Subscription`

### Layer 2: `lagentic-agentchat` — Chat Agents, Teams, LLM Integration
Built on core. Adds LLM awareness, tools, conversation history.

Key types: `ChatAgent` interface, `AssistantAgent`, `ChatCompletionClient` interface, `ChatCompletionContext` (Unbounded/Buffered/TokenLimited), `Workbench`, `Tool`/`FunctionTool`, `Team` interface, `BaseGroupChat`, `SelectorGroupChat`, `RoundRobinGroupChat`, `TerminationCondition` (composable And/Or)

Message types: `TextMessage`, `HandoffMessage`, `ToolCallMessage`, `ToolResultMessage`, `StopMessage`

### Package Dependency Rule (strict)
- `core` → stdlib only (never imports `agentchat` or `ext`)
- `agentchat` → `core` (never imports `ext`)
- `ext/*` providers → `agentchat`; `ext/glm`, `ext/deepseek` → `ext/openai`
- `ext/tools` → `core`
- `config` → `ext`, `agentchat`
- `cmd/lagentic` → `config`, `agentchat`

## 4-Agent Prototype

**Agents:** Coordinator (selector + task delegation), Coder (write code + fix), Reviewer (read-only review, no edits), Tester (write/run tests)

**Workbenches:** `toolWorkbench` (list_files, search_code → Coordinator), `coderWorkbench` (read/write_file, run_shell, search_code → Coder, Tester), `readOnlyWorkbench` (read_file, search_code → Reviewer)

**Orchestration:** `SelectorGroupChat` — Coordinator as LLM-based selector picks next speaker. Supports `HandoffMessage` for explicit agent handoffs.

**Termination:** `MaxTurnTermination(30).Or(TextMentionTermination("TASK COMPLETE"))`

## Multi-Provider Architecture

`Provider` interface with `ProviderRegistry` (providers self-register via `init()`). Built-in: Anthropic, OpenAI, GLM, DeepSeek, Ollama. OpenAI-compatible shortcut pattern: GLM/DeepSeek delegate to `openai.NewCompatibleClient` with custom `base_url`.

Adding a new provider: implement `Provider` interface (~50-100 lines for OpenAI-compatible), call `GlobalProviderRegistry.Register()` in `init()`, add config to `lagentic.yaml`. Zero framework changes.

Config file: `lagentic.yaml` — maps providers (api_key_env/base_url) and agents (model refs like `anthropic:claude-sonnet-4-6`).

## Error Handling

Sentinel errors: `ErrAgentNotFound`, `ErrProviderNotFound`, `ErrToolNotFound`, `ErrMaxTurnsExceeded`, `ErrContextCanceled`, `ErrTokenLimitExceeded`

Wrapped types: `AgentError{Agent, TaskID, Phase}`, `ProviderError{Provider, Model, Cause}`. Always wrap with context: `fmt.Errorf("doing X: %w", err)`.

## Build & Test Commands (once implemented)

```bash
go test ./...                                          # all unit tests
go test ./core/...                                     # single package
go test -race ./...                                    # race detector
LAGENTIC_INTEGRATION=1 go test ./ext/...               # integration tests (real API calls)
go test ./agentchat/ -run TestSelectorGroupChat_FeedbackLoop  # single test
```

## CLI Interface

```
lagentic                                      # interactive REPL
lagentic "implement a Go HTTP user service"   # one-shot mode
lagentic --config ./my-config.yaml            # custom config
lagentic --model deepseek:deepseek-coder      # override model
lagentic --verbose                            # show agent reasoning + tool calls
```

## Observability

`AgentEvent` stream: `speaker_selected`, `llm_call`, `llm_response`, `tool_call`, `tool_result`, `delegation`, `termination`. Display modes: normal (agent summary per turn), `--verbose` (tool calls, tokens), `--json` (structured events per line).

## Cancellation

`CancellationToken` flows from Ctrl+C through `MessageContext` to every agent and tool call. Long-running tools and LLM streams select on `ct.Done()`.
