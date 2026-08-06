package core

import (
	"testing"
	"time"
)

func TestCancellationToken_Done_BlocksBeforeCancel(t *testing.T) {
	ct := NewCancellationToken()
	select {
	case <-ct.Done():
		t.Error("Done() should block before Cancel()")
	default:
		// expected
	}
}

func TestCancellationToken_Done_ClosesAfterCancel(t *testing.T) {
	ct := NewCancellationToken()
	ct.Cancel()
	select {
	case <-ct.Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() should be closed after Cancel()")
	}
}

func TestCancellationToken_Cancel_Idempotent(t *testing.T) {
	ct := NewCancellationToken()
	ct.Cancel()
	ct.Cancel() // should not panic
	select {
	case <-ct.Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() should be closed after Cancel()")
	}
}

func TestMessageContext_Fields(t *testing.T) {
	ct := NewCancellationToken()
	mc := MessageContext{
		IsRPC:             true,
		TopicID:           TopicID{Type: "task", Source: "coord"},
		Sender:            AgentID{Type: "coordinator", Key: "team-1"},
		CancellationToken: ct,
	}
	if !mc.IsRPC {
		t.Error("IsRPC should be true")
	}
	if mc.TopicID.Type != "task" {
		t.Errorf("TopicID.Type = %q, want %q", mc.TopicID.Type, "task")
	}
	if mc.Sender.Type != "coordinator" {
		t.Errorf("Sender.Type = %q, want %q", mc.Sender.Type, "coordinator")
	}
	if mc.CancellationToken != ct {
		t.Error("CancellationToken should match")
	}
}
