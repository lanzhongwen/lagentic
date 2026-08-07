package agentchat

import (
	"context"
	"fmt"

	"github.com/lanzhongwen/lagentic/core"
)

// Team is a group of agents with orchestration.
type Team interface {
	Name() string
	Run(ctx context.Context, task string, ct *core.CancellationToken) (TaskResult, error)
	RunStream(ctx context.Context, task string, ct *core.CancellationToken) (<-chan AgentEvent, error)
	Reset(ctx context.Context) error
	SaveState() (map[string]any, error)
	LoadState(state map[string]any) error
}

// TaskResult is the output of a Team.Run call.
type TaskResult struct {
	Message    ChatMessage
	StopReason string
}

// GroupChatManager handles speaker selection and message thread.
type GroupChatManager interface {
	SelectSpeaker(thread []ChatMessage) ([]string, error)
	Reset() error
}

// BaseGroupChat is the foundation for all team types.
type BaseGroupChat struct {
	name         string
	participants []ChatAgent
	agentMap     map[string]ChatAgent
	manager      GroupChatManager
	termination  TerminationCondition
	thread       []ChatMessage
}

// NewBaseGroupChat creates a new BaseGroupChat.
func NewBaseGroupChat(name string, participants []ChatAgent, manager GroupChatManager, termination TerminationCondition) *BaseGroupChat {
	agentMap := make(map[string]ChatAgent, len(participants))
	for _, p := range participants {
		agentMap[p.Name()] = p
	}
	return &BaseGroupChat{
		name:         name,
		participants: participants,
		agentMap:     agentMap,
		manager:      manager,
		termination:  termination,
		thread:       make([]ChatMessage, 0),
	}
}

func (g *BaseGroupChat) Name() string { return g.name }

// Run executes the team until termination or error.
func (g *BaseGroupChat) Run(ctx context.Context, task string, ct *core.CancellationToken) (TaskResult, error) {
	// Seed the conversation with the user's task.
	g.thread = append(g.thread, NewTextMessage(task, "user"))

	for {
		// Check cancellation.
		if ct != nil {
			select {
			case <-ct.Done():
				return TaskResult{}, core.ErrContextCanceled
			default:
			}
		}

		// Check termination.
		if g.termination != nil {
			stop, err := g.termination.Check(g.thread)
			if err != nil {
				return TaskResult{}, fmt.Errorf("termination check: %w", err)
			}
			if stop != nil {
				return TaskResult{
					Message:    stop,
					StopReason: stop.Content,
				}, nil
			}
		}

		// Select next speaker.
		speakers, err := g.manager.SelectSpeaker(g.thread)
		if err != nil {
			return TaskResult{}, fmt.Errorf("select speaker: %w", err)
		}

		// Run each selected speaker.
		for _, speakerName := range speakers {
			agent, ok := g.agentMap[speakerName]
			if !ok {
				return TaskResult{}, fmt.Errorf("agent %q not found in team", speakerName)
			}

			resp, err := agent.OnMessages(ctx, g.thread, ct)
			if err != nil {
				return TaskResult{}, fmt.Errorf("agent %q: %w", speakerName, err)
			}

			msg := resp.ChatMessage
			g.thread = append(g.thread, msg)

			// If the agent produced a StopMessage, terminate immediately.
			if sm, isStop := msg.(StopMessage); isStop {
				return TaskResult{
					Message:    msg,
					StopReason: sm.Content,
				}, nil
			}

			// Re-check termination after each agent response.
			if g.termination != nil {
				stop, err := g.termination.Check(g.thread)
				if err != nil {
					return TaskResult{}, fmt.Errorf("termination check: %w", err)
				}
				if stop != nil {
					return TaskResult{
						Message:    stop,
						StopReason: stop.Content,
					}, nil
				}
			}
		}
	}
}

// RunStream executes the team and streams events.
func (g *BaseGroupChat) RunStream(ctx context.Context, task string, ct *core.CancellationToken) (<-chan AgentEvent, error) {
	ch := make(chan AgentEvent, 64)
	go func() {
		defer close(ch)
		g.thread = append(g.thread, NewTextMessage(task, "user"))
		ch <- AgentEvent{Type: "speaker_selected", Agent: "user", Data: task}

		for {
			if ct != nil {
				select {
				case <-ct.Done():
					ch <- AgentEvent{Type: "termination", Agent: "", Data: "canceled"}
					return
				default:
				}
			}

			if g.termination != nil {
				stop, err := g.termination.Check(g.thread)
				if err != nil || stop != nil {
					reason := "terminated"
					if stop != nil {
						reason = stop.Content
					}
					ch <- AgentEvent{Type: "termination", Agent: "", Data: reason}
					return
				}
			}

			speakers, err := g.manager.SelectSpeaker(g.thread)
			if err != nil {
				ch <- AgentEvent{Type: "termination", Agent: "", Data: err.Error()}
				return
			}

			for _, speakerName := range speakers {
				agent, ok := g.agentMap[speakerName]
				if !ok {
					ch <- AgentEvent{Type: "termination", Agent: "", Data: fmt.Sprintf("agent %q not found", speakerName)}
					return
				}

				ch <- AgentEvent{Type: "speaker_selected", Agent: speakerName, Data: nil}
				resp, err := agent.OnMessages(ctx, g.thread, ct)
				if err != nil {
					ch <- AgentEvent{Type: "termination", Agent: speakerName, Data: err.Error()}
					return
				}

				msg := resp.ChatMessage
				g.thread = append(g.thread, msg)
				ch <- AgentEvent{Type: "agent_response", Agent: speakerName, Data: msg}

				if sm, isStop := msg.(StopMessage); isStop {
					ch <- AgentEvent{Type: "termination", Agent: speakerName, Data: sm.Content}
					return
				}
			}
		}
	}()
	return ch, nil
}

func (g *BaseGroupChat) Reset(_ context.Context) error {
	g.thread = g.thread[:0]
	if g.manager != nil {
		if err := g.manager.Reset(); err != nil {
			return fmt.Errorf("reset manager: %w", err)
		}
	}
	if g.termination != nil {
		if err := g.termination.Reset(); err != nil {
			return fmt.Errorf("reset termination: %w", err)
		}
	}
	return nil
}

func (g *BaseGroupChat) SaveState() (map[string]any, error) {
	return map[string]any{
		"thread": g.thread,
	}, nil
}

func (g *BaseGroupChat) LoadState(state map[string]any) error {
	if thread, ok := state["thread"]; ok {
		g.thread = thread.([]ChatMessage)
	}
	return nil
}
