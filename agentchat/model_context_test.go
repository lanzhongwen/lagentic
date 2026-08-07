package agentchat

import "testing"

func TestUnboundedContext_AddAndGetMessages(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	msg1 := LLMMessage{Role: LLMMessageRoleUser, Content: "hello"}
	msg2 := LLMMessage{Role: LLMMessageRoleAssistant, Content: "hi"}

	if err := ctx.AddMessage(msg1); err != nil {
		t.Fatalf("AddMessage() error: %v", err)
	}
	if err := ctx.AddMessage(msg2); err != nil {
		t.Fatalf("AddMessage() error: %v", err)
	}

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("msgs[0].Content = %q, want %q", msgs[0].Content, "hello")
	}
	if msgs[1].Content != "hi" {
		t.Errorf("msgs[1].Content = %q, want %q", msgs[1].Content, "hi")
	}
}

func TestUnboundedContext_Clear(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "hello"})

	if err := ctx.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}
	if len(ctx.GetMessages()) != 0 {
		t.Errorf("len(GetMessages()) after Clear = %d, want 0", len(ctx.GetMessages()))
	}
}

func TestUnboundedContext_SaveLoadState(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "hello"})

	state, err := ctx.SaveState()
	if err != nil {
		t.Fatalf("SaveState() error: %v", err)
	}

	ctx2 := NewUnboundedChatCompletionContext()
	if err := ctx2.LoadState(state); err != nil {
		t.Fatalf("LoadState() error: %v", err)
	}
	if len(ctx2.GetMessages()) != 1 {
		t.Fatalf("len(GetMessages()) after LoadState = %d, want 1", len(ctx2.GetMessages()))
	}
	if ctx2.GetMessages()[0].Content != "hello" {
		t.Errorf("Content after LoadState = %q, want %q", ctx2.GetMessages()[0].Content, "hello")
	}
}

func TestUnboundedContext_Interface(t *testing.T) {
	var _ ChatCompletionContext = NewUnboundedChatCompletionContext()
}

func TestBufferedContext_TruncatesToLimit(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(2)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg2"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg3"})

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
	// Should keep the last 2 messages
	if msgs[0].Content != "msg2" {
		t.Errorf("msgs[0].Content = %q, want %q", msgs[0].Content, "msg2")
	}
	if msgs[1].Content != "msg3" {
		t.Errorf("msgs[1].Content = %q, want %q", msgs[1].Content, "msg3")
	}
}

func TestBufferedContext_UnderLimit_NoTruncation(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(5)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg2"})

	msgs := ctx.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("len(GetMessages()) = %d, want 2", len(msgs))
	}
}

func TestBufferedContext_Clear(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(5)
	ctx.AddMessage(LLMMessage{Role: LLMMessageRoleUser, Content: "msg1"})
	ctx.Clear()
	if len(ctx.GetMessages()) != 0 {
		t.Errorf("len after Clear = %d, want 0", len(ctx.GetMessages()))
	}
}

func TestBufferedContext_Interface(t *testing.T) {
	var _ ChatCompletionContext = NewBufferedChatCompletionContext(10)
}

func TestUnboundedContext_LoadState_InvalidData(t *testing.T) {
	ctx := NewUnboundedChatCompletionContext()
	err := ctx.LoadState(map[string]any{"messages": "not a slice"})
	if err == nil {
		t.Error("expected error for invalid messages type")
	}
}

func TestBufferedContext_LoadState_InvalidData(t *testing.T) {
	ctx := NewBufferedChatCompletionContext(5)
	err := ctx.LoadState(map[string]any{"messages": "not a slice"})
	if err == nil {
		t.Error("expected error for invalid messages type")
	}
}

func TestBufferedContext_InvalidLimit_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for zero limit")
		}
	}()
	NewBufferedChatCompletionContext(0)
}
