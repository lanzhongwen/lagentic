package agentchat

import "fmt"

// TerminationCondition decides when a team should stop. Composable via And()/Or().
type TerminationCondition interface {
	Check(messages []ChatMessage) (*StopMessage, error)
	Reset() error
}

// MaxTurnTermination stops after a maximum number of messages.
type MaxTurnTermination int

func (m MaxTurnTermination) Check(messages []ChatMessage) (*StopMessage, error) {
	if len(messages) >= int(m) {
		msg := NewStopMessage(fmt.Sprintf("Maximum number of turns %d reached", int(m)), "MaxTurnTermination")
		return &msg, nil
	}
	return nil, nil
}

func (m MaxTurnTermination) Reset() error { return nil }

// TextMentionTermination stops when a message contains the specified text.
type TextMentionTermination string

func (t TextMentionTermination) Check(messages []ChatMessage) (*StopMessage, error) {
	text := string(t)
	for _, msg := range messages {
		if tm, ok := msg.(TextMessage); ok && tm.Content == text {
			sm := NewStopMessage(fmt.Sprintf("Text %q mentioned", text), "TextMentionTermination")
			return &sm, nil
		}
	}
	return nil, nil
}

func (t TextMentionTermination) Reset() error { return nil }

// AndTermination stops when both conditions are met.
type AndTermination struct {
	Left  TerminationCondition
	Right TerminationCondition
}

func (a *AndTermination) Check(messages []ChatMessage) (*StopMessage, error) {
	left, err := a.Left.Check(messages)
	if err != nil {
		return nil, fmt.Errorf("and left: %w", err)
	}
	right, err := a.Right.Check(messages)
	if err != nil {
		return nil, fmt.Errorf("and right: %w", err)
	}
	if left != nil && right != nil {
		return left, nil
	}
	return nil, nil
}

func (a *AndTermination) Reset() error {
	if err := a.Left.Reset(); err != nil {
		return err
	}
	return a.Right.Reset()
}

// OrTermination stops when either condition is met.
type OrTermination struct {
	Left  TerminationCondition
	Right TerminationCondition
}

func (o *OrTermination) Check(messages []ChatMessage) (*StopMessage, error) {
	left, err := o.Left.Check(messages)
	if err != nil {
		return nil, fmt.Errorf("or left: %w", err)
	}
	if left != nil {
		return left, nil
	}
	right, err := o.Right.Check(messages)
	if err != nil {
		return nil, fmt.Errorf("or right: %w", err)
	}
	return right, nil
}

func (o *OrTermination) Reset() error {
	if err := o.Left.Reset(); err != nil {
		return err
	}
	return o.Right.Reset()
}

// And composes two conditions — both must be met to stop.
func (t MaxTurnTermination) And(other TerminationCondition) TerminationCondition {
	return &AndTermination{Left: t, Right: other}
}

// Or composes two conditions — either can trigger a stop.
func (t MaxTurnTermination) Or(other TerminationCondition) TerminationCondition {
	return &OrTermination{Left: t, Right: other}
}

// And composes two conditions on TextMentionTermination.
func (t TextMentionTermination) And(other TerminationCondition) TerminationCondition {
	return &AndTermination{Left: t, Right: other}
}

// Or composes two conditions on TextMentionTermination.
func (t TextMentionTermination) Or(other TerminationCondition) TerminationCondition {
	return &OrTermination{Left: t, Right: other}
}
