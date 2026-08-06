package core

import "testing"

func TestAgentID_String_Format(t *testing.T) {
	id := AgentID{Type: "coordinator", Key: "team-1"}
	want := "coordinator/team-1"
	if got := id.String(); got != want {
		t.Errorf("AgentID.String() = %q, want %q", got, want)
	}
}

func TestAgentID_String_Empty(t *testing.T) {
	id := AgentID{}
	want := "/"
	if got := id.String(); got != want {
		t.Errorf("AgentID.String() = %q, want %q", got, want)
	}
}

func TestAgentID_Equal(t *testing.T) {
	a := AgentID{Type: "coder", Key: "default"}
	b := AgentID{Type: "coder", Key: "default"}
	c := AgentID{Type: "reviewer", Key: "default"}
	if a != b {
		t.Error("identical AgentIDs should be equal")
	}
	if a == c {
		t.Error("different AgentIDs should not be equal")
	}
}
