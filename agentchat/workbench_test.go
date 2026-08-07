package agentchat

import (
	"context"
	"encoding/json"
	"errors"
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
	// The tool returns string(args) = `{"msg":"hi"}`, which json.Marshal encodes as a JSON string.
	want := `"{\"msg\":\"hi\"}"`
	if result.Content != want {
		t.Errorf("result.Content = %q, want %q", result.Content, want)
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
	if !errors.Is(err, core.ErrToolNotFound) {
		t.Errorf("error = %v, want wrapped core.ErrToolNotFound", err)
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
	if result.Content != `"v2"` {
		t.Errorf("result.Content = %q, want %q", result.Content, `"v2"`)
	}
}

func TestWorkbench_Interface(t *testing.T) {
	var _ Workbench = NewStaticWorkbench()
}
