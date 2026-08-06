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

// Compile-time interface check.
var _ Tool = (*FunctionTool)(nil)

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
