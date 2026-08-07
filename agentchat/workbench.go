package agentchat

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lanzhongwen/lagentic/core"
)

// Workbench is a container for tools available to an agent.
type Workbench interface {
	ListTools() []core.ToolSchema
	CallTool(ctx context.Context, name string, args json.RawMessage, ct *core.CancellationToken) (core.ToolResult, error)
	Register(tool core.Tool) error
}

// StaticWorkbench is a simple in-memory workbench.
type StaticWorkbench struct {
	mu    sync.RWMutex
	tools map[string]core.Tool
}

// NewStaticWorkbench creates an empty workbench.
func NewStaticWorkbench() *StaticWorkbench {
	return &StaticWorkbench{
		tools: make(map[string]core.Tool),
	}
}

func (w *StaticWorkbench) Register(tool core.Tool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tools[tool.Name()] = tool
	return nil
}

func (w *StaticWorkbench) ListTools() []core.ToolSchema {
	w.mu.RLock()
	defer w.mu.RUnlock()
	schemas := make([]core.ToolSchema, 0, len(w.tools))
	for _, tool := range w.tools {
		schemas = append(schemas, tool.Schema())
	}
	return schemas
}

func (w *StaticWorkbench) CallTool(ctx context.Context, name string, args json.RawMessage, ct *core.CancellationToken) (core.ToolResult, error) {
	w.mu.RLock()
	tool, ok := w.tools[name]
	w.mu.RUnlock()
	if !ok {
		return core.ToolResult{}, fmt.Errorf("workbench: %w", core.ErrToolNotFound)
	}

	result, err := tool.RunJSON(ctx, args, ct)
	if err != nil {
		return core.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	return core.ToolResult{Content: fmt.Sprintf("%v", result)}, nil
}
