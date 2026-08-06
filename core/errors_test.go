package core

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
		{"AgentNotFound", ErrAgentNotFound, "agent not found"},
		{"ToolNotFound", ErrToolNotFound, "tool not found"},
		{"ContextCanceled", ErrContextCanceled, "context canceled"},
		{"NoHandler", ErrNoHandler, "no handler registered"},
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
