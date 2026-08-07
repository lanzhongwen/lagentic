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
	source  string
}

func (m TextMessage) Type() string  { return "TextMessage" }
func (m TextMessage) Source() string { return m.source }

// NewTextMessage creates a TextMessage with the given content and source.
func NewTextMessage(content, source string) TextMessage {
	return TextMessage{Content: content, source: source}
}

// HandoffMessage explicitly transfers control to another agent.
type HandoffMessage struct {
	Target  string
	Context string
	source  string
}

func (m HandoffMessage) Type() string  { return "HandoffMessage" }
func (m HandoffMessage) Source() string { return m.source }

// NewHandoffMessage creates a HandoffMessage with the given target, context, and source.
func NewHandoffMessage(target, context, source string) HandoffMessage {
	return HandoffMessage{Target: target, Context: context, source: source}
}

// ToolCall represents a single tool invocation request.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // JSON-encoded arguments
}

// ToolCallMessage carries one or more tool call requests.
type ToolCallMessage struct {
	ToolCalls []ToolCall
	source    string
}

func (m ToolCallMessage) Type() string  { return "ToolCallMessage" }
func (m ToolCallMessage) Source() string { return m.source }

// NewToolCallMessage creates a ToolCallMessage with the given tool calls and source.
func NewToolCallMessage(toolCalls []ToolCall, source string) ToolCallMessage {
	return ToolCallMessage{ToolCalls: toolCalls, source: source}
}

// ToolResultMessage carries the results of tool executions.
type ToolResultMessage struct {
	Results []core.ToolResult
	source  string
}

func (m ToolResultMessage) Type() string  { return "ToolResultMessage" }
func (m ToolResultMessage) Source() string { return m.source }

// NewToolResultMessage creates a ToolResultMessage with the given results and source.
func NewToolResultMessage(results []core.ToolResult, source string) ToolResultMessage {
	return ToolResultMessage{Results: results, source: source}
}

// StopMessage signals that the agent or team should stop.
type StopMessage struct {
	Content string
	source  string
}

func (m StopMessage) Type() string  { return "StopMessage" }
func (m StopMessage) Source() string { return m.source }

// NewStopMessage creates a StopMessage with the given content and source.
func NewStopMessage(content, source string) StopMessage {
	return StopMessage{Content: content, source: source}
}
