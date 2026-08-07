package agentchat

// Handoff defines a potential transfer of control to another agent.
type Handoff struct {
	Target      string
	Description string
}

// ToHandoffMessage creates a HandoffMessage from this handoff definition.
func (h Handoff) ToHandoffMessage(source string) HandoffMessage {
	return NewHandoffMessage(h.Target, h.Description, source)
}
