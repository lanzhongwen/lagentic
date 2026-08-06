package core

import "fmt"

// AgentID uniquely identifies an agent instance.
type AgentID struct {
	Type string // agent type (e.g., "coordinator", "coder")
	Key  string // instance key (e.g., team UUID)
}

func (id AgentID) String() string {
	return fmt.Sprintf("%s/%s", id.Type, id.Key)
}
