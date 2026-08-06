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
