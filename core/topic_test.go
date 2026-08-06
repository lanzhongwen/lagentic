package core

import "testing"

func TestTopicID_String_Format(t *testing.T) {
	id := TopicID{Type: "task", Source: "coordinator"}
	want := "task/coordinator"
	if got := id.String(); got != want {
		t.Errorf("TopicID.String() = %q, want %q", got, want)
	}
}

func TestTopicID_Equal(t *testing.T) {
	a := TopicID{Type: "task", Source: "coord"}
	b := TopicID{Type: "task", Source: "coord"}
	c := TopicID{Type: "result", Source: "coord"}
	if a != b {
		t.Error("identical TopicIDs should be equal")
	}
	if a == c {
		t.Error("different TopicIDs should not be equal")
	}
}
