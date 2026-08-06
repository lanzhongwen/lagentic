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

func TestTypeSubscription_IsMatch_MatchingType(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "task", Source: "coordinator"}
	if !sub.IsMatch(topic) {
		t.Error("should match when topic type equals subscription type")
	}
}

func TestTypeSubscription_IsMatch_NonMatchingType(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "result", Source: "coordinator"}
	if sub.IsMatch(topic) {
		t.Error("should not match when topic type differs")
	}
}

func TestTypeSubscription_MapToAgent(t *testing.T) {
	sub := NewTypeSubscription("task", "coordinator")
	topic := TopicID{Type: "task", Source: "any"}
	got := sub.MapToAgent(topic)
	want := AgentID{Type: "coordinator", Key: "default"}
	if got != want {
		t.Errorf("MapToAgent() = %v, want %v", got, want)
	}
}
