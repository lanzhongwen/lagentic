package core

import "fmt"

// Sentinel errors for distinguishable failure modes.
var (
	ErrAgentNotFound   = fmt.Errorf("agent not found")
	ErrToolNotFound    = fmt.Errorf("tool not found")
	ErrContextCanceled = fmt.Errorf("context canceled")
	ErrNoHandler       = fmt.Errorf("no handler registered")
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
