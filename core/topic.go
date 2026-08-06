package core

import "fmt"

// TopicID identifies a pub/sub topic.
type TopicID struct {
	Type   string // topic type
	Source string // namespace/context
}

func (id TopicID) String() string {
	return fmt.Sprintf("%s/%s", id.Type, id.Source)
}

// Subscription maps topics to agents.
type Subscription interface {
	IsMatch(topic TopicID) bool
	MapToAgent(topic TopicID) AgentID
}

// TypeSubscription matches topics by Type and maps them to a specific agent type.
type TypeSubscription struct {
	topicType string
	agentType string
	agentKey  string
}

// NewTypeSubscription creates a subscription that matches topics with the given type
// and maps them to an agent of the given type and key.
func NewTypeSubscription(topicType, agentType string, agentKey ...string) *TypeSubscription {
	key := "default"
	if len(agentKey) > 0 {
		key = agentKey[0]
	}
	return &TypeSubscription{
		topicType: topicType,
		agentType: agentType,
		agentKey:  key,
	}
}

func (s *TypeSubscription) IsMatch(topic TopicID) bool {
	return topic.Type == s.topicType
}

func (s *TypeSubscription) MapToAgent(topic TopicID) AgentID {
	return AgentID{Type: s.agentType, Key: s.agentKey}
}
