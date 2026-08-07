package agentchat

import "testing"

func TestHandoff_Fields(t *testing.T) {
	h := Handoff{Target: "reviewer", Description: "Code is ready for review"}
	if h.Target != "reviewer" {
		t.Errorf("Target = %q, want %q", h.Target, "reviewer")
	}
	if h.Description != "Code is ready for review" {
		t.Errorf("Description = %q, want %q", h.Description, "Code is ready for review")
	}
}

func TestHandoff_ProducesHandoffMessage(t *testing.T) {
	h := Handoff{Target: "reviewer", Description: "Code is ready"}
	msg := h.ToHandoffMessage("coder")
	if msg.Target != "reviewer" {
		t.Errorf("Target = %q, want %q", msg.Target, "reviewer")
	}
	if msg.Context != "Code is ready" {
		t.Errorf("Context = %q, want %q", msg.Context, "Code is ready")
	}
	if msg.Source() != "coder" {
		t.Errorf("Source() = %q, want %q", msg.Source(), "coder")
	}
}
