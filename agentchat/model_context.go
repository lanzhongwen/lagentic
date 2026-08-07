package agentchat

import (
	"fmt"
	"sync"
)

// ChatCompletionContext manages per-agent conversation history.
// Note: SaveState/LoadState use map[string]any with Go type assertions,
// suitable for in-memory state transfer only. JSON serialization is not
// supported — a persistence-compatible format will be added in a future iteration.
type ChatCompletionContext interface {
	AddMessage(msg LLMMessage) error
	GetMessages() []LLMMessage
	Clear() error
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
}

// UnboundedChatCompletionContext stores messages without limit.
type UnboundedChatCompletionContext struct {
	mu       sync.RWMutex
	messages []LLMMessage
}

// NewUnboundedChatCompletionContext creates an unbounded context.
func NewUnboundedChatCompletionContext() *UnboundedChatCompletionContext {
	return &UnboundedChatCompletionContext{
		messages: make([]LLMMessage, 0),
	}
}

func (c *UnboundedChatCompletionContext) AddMessage(msg LLMMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, msg)
	return nil
}

func (c *UnboundedChatCompletionContext) GetMessages() []LLMMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messages
}

func (c *UnboundedChatCompletionContext) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = nil
	return nil
}

func (c *UnboundedChatCompletionContext) SaveState() (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]any{
		"messages": c.messages,
	}, nil
}

func (c *UnboundedChatCompletionContext) LoadState(state map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if msgs, ok := state["messages"]; ok {
		typed, ok := msgs.([]LLMMessage)
		if !ok {
			return fmt.Errorf("LoadState: invalid \"messages\" type %T", msgs)
		}
		c.messages = typed
	}
	return nil
}

// BufferedChatCompletionContext keeps only the last N messages.
type BufferedChatCompletionContext struct {
	mu       sync.RWMutex
	messages []LLMMessage
	limit    int
}

// NewBufferedChatCompletionContext creates a context that retains at most limit messages.
func NewBufferedChatCompletionContext(limit int) *BufferedChatCompletionContext {
	if limit <= 0 {
		panic(fmt.Sprintf("BufferedChatCompletionContext: limit must be positive, got %d", limit))
	}
	return &BufferedChatCompletionContext{
		messages: make([]LLMMessage, 0),
		limit:    limit,
	}
}

func (c *BufferedChatCompletionContext) AddMessage(msg LLMMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, msg)
	if len(c.messages) > c.limit {
		c.messages = c.messages[len(c.messages)-c.limit:]
	}
	return nil
}

func (c *BufferedChatCompletionContext) GetMessages() []LLMMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messages
}

func (c *BufferedChatCompletionContext) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = nil
	return nil
}

func (c *BufferedChatCompletionContext) SaveState() (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]any{
		"messages": c.messages,
		"limit":    c.limit,
	}, nil
}

func (c *BufferedChatCompletionContext) LoadState(state map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if msgs, ok := state["messages"]; ok {
		typed, ok := msgs.([]LLMMessage)
		if !ok {
			return fmt.Errorf("LoadState: invalid \"messages\" type %T", msgs)
		}
		c.messages = typed
	}
	if limit, ok := state["limit"]; ok {
		typed, ok := limit.(int)
		if !ok {
			return fmt.Errorf("LoadState: invalid \"limit\" type %T", limit)
		}
		c.limit = typed
	}
	return nil
}
