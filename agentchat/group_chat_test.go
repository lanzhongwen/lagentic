package agentchat

import (
	"context"
	"errors"
	"testing"

	"github.com/lanzhongwen/lagentic/core"
)

// groupMockChatAgent implements ChatAgent for group chat testing.
type groupMockChatAgent struct {
	name        string
	description string
	responses   []Response
	callCount   int
}

func (a *groupMockChatAgent) Name() string        { return a.name }
func (a *groupMockChatAgent) Description() string  { return a.description }
func (a *groupMockChatAgent) OnMessages(_ context.Context, _ []ChatMessage, _ *core.CancellationToken) (Response, error) {
	if a.callCount >= len(a.responses) {
		return Response{ChatMessage: NewTextMessage("default response", a.name)}, nil
	}
	resp := a.responses[a.callCount]
	a.callCount++
	return resp, nil
}
func (a *groupMockChatAgent) OnMessagesStream(_ context.Context, _ []ChatMessage, _ *core.CancellationToken) (<-chan AgentEvent, error) {
	ch := make(chan AgentEvent, 1)
	close(ch)
	return ch, nil
}
func (a *groupMockChatAgent) OnReset(_ context.Context) error { a.callCount = 0; return nil }
func (a *groupMockChatAgent) SaveState() (map[string]any, error) { return nil, nil }
func (a *groupMockChatAgent) LoadState(_ map[string]any) error    { return nil }
func (a *groupMockChatAgent) Close() error                        { return nil }

// groupMockManager implements GroupChatManager for group chat testing.
type groupMockManager struct {
	speakers []string
	index    int
}

func (m *groupMockManager) Reset() error { m.index = 0; return nil }

func (m *groupMockManager) SelectSpeaker(_ []ChatMessage) ([]string, error) {
	if m.index >= len(m.speakers) {
		return nil, ErrMaxTurnsExceeded
	}
	speaker := m.speakers[m.index]
	m.index++
	return []string{speaker}, nil
}

func TestBaseGroupChat_Run_SingleRoundTrip(t *testing.T) {
	agent := &groupMockChatAgent{
		name:      "echo",
		responses: []Response{{ChatMessage: NewTextMessage("hello back", "echo")}},
	}
	manager := &groupMockManager{speakers: []string{"echo"}}
	termination := MaxTurnTermination(2)

	team := NewBaseGroupChat(
		"test-team",
		[]ChatAgent{agent},
		manager,
		termination,
	)

	result, err := team.Run(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.Message == nil {
		t.Fatal("result.Message should not be nil")
	}
}

func TestBaseGroupChat_Run_MaxTurnsExceeded(t *testing.T) {
	agent := &groupMockChatAgent{
		name:      "chatty",
		responses: []Response{{ChatMessage: NewTextMessage("blah", "chatty")}},
	}
	// Manager only has 2 speakers, so the 3rd SelectSpeaker call returns ErrMaxTurnsExceeded.
	manager := &groupMockManager{speakers: []string{"chatty", "chatty"}}
	// High turn limit so the manager exhausts first (1 user + 2 agent = 3 messages < 10).
	termination := MaxTurnTermination(10)

	team := NewBaseGroupChat(
		"test-team",
		[]ChatAgent{agent},
		manager,
		termination,
	)

	_, err := team.Run(context.Background(), "start", nil)
	if err == nil {
		t.Fatal("expected error when max turns exceeded")
	}
	if !errors.Is(err, ErrMaxTurnsExceeded) {
		t.Errorf("error = %v, want ErrMaxTurnsExceeded", err)
	}
}

func TestBaseGroupChat_Run_StopMessage(t *testing.T) {
	agent := &groupMockChatAgent{
		name: "stopper",
		responses: []Response{
			{ChatMessage: NewStopMessage("done", "stopper")},
		},
	}
	manager := &groupMockManager{speakers: []string{"stopper"}}
	termination := MaxTurnTermination(10)

	team := NewBaseGroupChat(
		"test-team",
		[]ChatAgent{agent},
		manager,
		termination,
	)

	result, err := team.Run(context.Background(), "start", nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.StopReason == "" {
		t.Error("expected non-empty StopReason")
	}
}

func TestBaseGroupChat_Reset(t *testing.T) {
	agent := &groupMockChatAgent{name: "a"}
	manager := &groupMockManager{speakers: []string{"a"}}
	termination := MaxTurnTermination(5)
	team := NewBaseGroupChat("team", []ChatAgent{agent}, manager, termination)

	if err := team.Reset(context.Background()); err != nil {
		t.Fatalf("Reset() error: %v", err)
	}
}

func TestBaseGroupChat_Name(t *testing.T) {
	team := NewBaseGroupChat("my-team", nil, nil, nil)
	if team.Name() != "my-team" {
		t.Errorf("Name() = %q, want %q", team.Name(), "my-team")
	}
}

func TestTeam_Interface(t *testing.T) {
	var _ Team = NewBaseGroupChat("team", nil, nil, nil)
}
