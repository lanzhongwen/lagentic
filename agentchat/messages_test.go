package agentchat

import (
	"testing"

	"github.com/lanzhongwen/lagentic/core"
)

func TestTextMessage_Type(t *testing.T) {
	msg := NewTextMessage("hello", "coder")
	if msg.Type() != "TextMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "TextMessage")
	}
}

func TestTextMessage_Source(t *testing.T) {
	msg := NewTextMessage("hello", "coder")
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestTextMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = NewTextMessage("hello", "coder")
}

func TestHandoffMessage_Type(t *testing.T) {
	msg := NewHandoffMessage("reviewer", "code ready", "coder")
	if msg.Type() != "HandoffMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "HandoffMessage")
	}
}

func TestHandoffMessage_Source(t *testing.T) {
	msg := NewHandoffMessage("reviewer", "code ready", "coder")
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestHandoffMessage_Fields(t *testing.T) {
	msg := NewHandoffMessage("reviewer", "code ready", "coder")
	if msg.Target != "reviewer" {
		t.Errorf("Target = %q, want %q", msg.Target, "reviewer")
	}
	if msg.Context != "code ready" {
		t.Errorf("Context = %q, want %q", msg.Context, "code ready")
	}
}

func TestHandoffMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = NewHandoffMessage("reviewer", "code ready", "coder")
}

func TestToolCallMessage_Type(t *testing.T) {
	msg := NewToolCallMessage(
		[]ToolCall{{ID: "tc1", Name: "read_file", Arguments: `{"path":"main.go"}`}},
		"coder",
	)
	if msg.Type() != "ToolCallMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "ToolCallMessage")
	}
}

func TestToolCallMessage_Source(t *testing.T) {
	msg := NewToolCallMessage(
		[]ToolCall{{ID: "tc1", Name: "read_file", Arguments: `{"path":"main.go"}`}},
		"coder",
	)
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestToolCallMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = NewToolCallMessage(nil, "coder")
}

func TestToolResultMessage_Type(t *testing.T) {
	msg := NewToolResultMessage(
		[]core.ToolResult{{Content: "file contents", IsError: false}},
		"coder",
	)
	if msg.Type() != "ToolResultMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "ToolResultMessage")
	}
}

func TestToolResultMessage_Source(t *testing.T) {
	msg := NewToolResultMessage(
		[]core.ToolResult{{Content: "file contents", IsError: false}},
		"coder",
	)
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}

func TestToolResultMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = NewToolResultMessage(nil, "coder")
}

func TestStopMessage_Type(t *testing.T) {
	msg := NewStopMessage("task complete", "coordinator")
	if msg.Type() != "StopMessage" {
		t.Errorf("Type() = %q, want %q", msg.Type(), "StopMessage")
	}
}

func TestStopMessage_Source(t *testing.T) {
	msg := NewStopMessage("task complete", "coordinator")
	if msg.Source() != "coordinator" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coordinator")
	}
}

func TestStopMessage_BaseChatMessage(t *testing.T) {
	var _ BaseChatMessage = NewStopMessage("task complete", "coordinator")
}

func TestToolCall_Fields(t *testing.T) {
	tc := ToolCall{ID: "tc1", Name: "read_file", Arguments: `{"path":"main.go"}`}
	if tc.ID != "tc1" {
		t.Errorf("ID = %q, want %q", tc.ID, "tc1")
	}
	if tc.Name != "read_file" {
		t.Errorf("Name = %q, want %q", tc.Name, "read_file")
	}
	if tc.Arguments != `{"path":"main.go"}` {
		t.Errorf("Arguments = %q, want %q", tc.Arguments, `{"path":"main.go"}`)
	}
}

func TestBaseChatMessage_Interface_AllTypes(t *testing.T) {
	msgs := []BaseChatMessage{
		NewTextMessage("hello", "coder"),
		NewHandoffMessage("reviewer", "done", "coder"),
		NewToolCallMessage([]ToolCall{{ID: "1", Name: "f"}}, "coder"),
		NewToolResultMessage([]core.ToolResult{{Content: "ok"}}, "coder"),
		NewStopMessage("done", "coordinator"),
	}
	wantTypes := []string{"TextMessage", "HandoffMessage", "ToolCallMessage", "ToolResultMessage", "StopMessage"}
	for i, msg := range msgs {
		if msg.Type() != wantTypes[i] {
			t.Errorf("msg[%d].Type() = %q, want %q", i, msg.Type(), wantTypes[i])
		}
	}
}
