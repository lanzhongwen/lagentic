package agentchat

import "fmt"

// ChatCompletionContext manages per-agent conversation history.
type ChatCompletionContext interface {
	AddMessage(msg LLMMessage) error
	GetMessages() []LLMMessage
	Clear() error
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
}

// UnboundedChatCompletionContext stores messages without limit.
type UnboundedChatCompletionContext struct {
	messages []LLMMessage
}

// NewUnboundedChatCompletionContext creates an unbounded context.
func NewUnboundedChatCompletionContext() *UnboundedChatCompletionContext {
	return &UnboundedChatCompletionContext{
		messages: make([]LLMMessage, 0),
	}
}

func (c *UnboundedChatCompletionContext) AddMessage(msg LLMMessage) error {
	c.messages = append(c.messages, msg)
	return nil
}

func (c *UnboundedChatCompletionContext) GetMessages() []LLMMessage {
	return c.messages
}

func (c *UnboundedChatCompletionContext) Clear() error {
	c.messages = c.messages[:0]
	return nil
}

func (c *UnboundedChatCompletionContext) SaveState() (map[string]any, error) {
	return map[string]any{
		"messages": c.messages,
	}, nil
}

func (c *UnboundedChatCompletionContext) LoadState(state map[string]any) error {
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
	c.messages = append(c.messages, msg)
	if len(c.messages) > c.limit {
		c.messages = c.messages[len(c.messages)-c.limit:]
	}
	return nil
}

func (c *BufferedChatCompletionContext) GetMessages() []LLMMessage {
	return c.messages
}

func (c *BufferedChatCompletionContext) Clear() error {
	c.messages = c.messages[:0]
	return nil
}

func (c *BufferedChatCompletionContext) SaveState() (map[string]any, error) {
	return map[string]any{
		"messages": c.messages,
		"limit":    c.limit,
	}, nil
}

func (c *BufferedChatCompletionContext) LoadState(state map[string]any) error {
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
