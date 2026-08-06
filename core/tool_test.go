package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
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
