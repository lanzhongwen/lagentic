package agentchat

import "testing"

func TestMaxTurnTermination_Check_BelowMax_ReturnsNil(t *testing.T) {
	tc := MaxTurnTermination(3)
	msgs := []ChatMessage{
		NewTextMessage("turn 1", "a"),
		NewTextMessage("turn 2", "b"),
	}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop != nil {
		t.Errorf("stop = %v, want nil (only 2 messages, max is 3)", stop)
	}
}

func TestMaxTurnTermination_Check_AtMax_ReturnsStop(t *testing.T) {
	tc := MaxTurnTermination(3)
	msgs := []ChatMessage{
		NewTextMessage("turn 1", "a"),
		NewTextMessage("turn 2", "b"),
		NewTextMessage("turn 3", "c"),
	}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop == nil {
		t.Fatal("expected StopMessage at max turns")
	}
	if stop.Content != "Maximum number of turns 3 reached" {
		t.Errorf("Content = %q, want max turns message", stop.Content)
	}
}

func TestMaxTurnTermination_Reset(t *testing.T) {
	tc := MaxTurnTermination(2)
	msgs := []ChatMessage{NewTextMessage("t1", "a"), NewTextMessage("t2", "b")}
	tc.Check(msgs)
	if err := tc.Reset(); err != nil {
		t.Fatalf("Reset() error: %v", err)
	}
	// After reset, should be able to check again from scratch
	stop, _ := tc.Check([]ChatMessage{NewTextMessage("t1", "a")})
	if stop != nil {
		t.Error("after reset, 1 message should not exceed max of 2")
	}
}

func TestTextMentionTermination_Check_NoMatch_ReturnsNil(t *testing.T) {
	tc := TextMentionTermination("TASK COMPLETE")
	msgs := []ChatMessage{NewTextMessage("still working", "coder")}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop != nil {
		t.Errorf("stop = %v, want nil (no match)", stop)
	}
}

func TestTextMentionTermination_Check_Match_ReturnsStop(t *testing.T) {
	tc := TextMentionTermination("TASK COMPLETE")
	msgs := []ChatMessage{NewTextMessage("TASK COMPLETE", "coordinator")}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop == nil {
		t.Fatal("expected StopMessage when text matches")
	}
}

func TestTextMentionTermination_Reset(t *testing.T) {
	tc := TextMentionTermination("DONE")
	tc.Check([]ChatMessage{NewTextMessage("DONE", "a")})
	if err := tc.Reset(); err != nil {
		t.Fatalf("Reset() error: %v", err)
	}
}

func TestAndTermination_BothMet_ReturnsStop(t *testing.T) {
	tc := MaxTurnTermination(2).And(TextMentionTermination("DONE"))
	// Both: 2 messages AND text "DONE"
	msgs := []ChatMessage{
		NewTextMessage("work", "a"),
		NewTextMessage("DONE", "b"),
	}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop == nil {
		t.Fatal("expected StopMessage when both conditions met")
	}
}

func TestAndTermination_OneMet_ReturnsNil(t *testing.T) {
	tc := MaxTurnTermination(2).And(TextMentionTermination("DONE"))
	// Max turns met but text not found
	msgs := []ChatMessage{
		NewTextMessage("work", "a"),
		NewTextMessage("still working", "b"),
	}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop != nil {
		t.Error("expected nil when only one condition met")
	}
}

func TestOrTermination_EitherMet_ReturnsStop(t *testing.T) {
	tc := MaxTurnTermination(1).Or(TextMentionTermination("DONE"))
	// Max turns met (1), text not found
	msgs := []ChatMessage{NewTextMessage("work", "a")}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop == nil {
		t.Fatal("expected StopMessage when either condition met")
	}
}

func TestOrTermination_NeitherMet_ReturnsNil(t *testing.T) {
	tc := MaxTurnTermination(5).Or(TextMentionTermination("DONE"))
	msgs := []ChatMessage{NewTextMessage("work", "a")}
	stop, err := tc.Check(msgs)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if stop != nil {
		t.Error("expected nil when neither condition met")
	}
}

func TestTerminationCondition_Interface(t *testing.T) {
	var _ TerminationCondition = MaxTurnTermination(1)
	var _ TerminationCondition = TextMentionTermination("x")
}
