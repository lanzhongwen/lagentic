package agentchat

// GroupChatStart is the initial event sent when a GroupChat begins.
type GroupChatStart struct {
	Messages []ChatMessage
}

// GroupChatRequestPublish signals that an agent should publish a response.
type GroupChatRequestPublish struct{}

// GroupChatAgentResponse carries an agent's response within the GroupChat.
type GroupChatAgentResponse struct {
	Agent    ChatAgent
	Response Response
}

// GroupChatTermination signals that the GroupChat has terminated.
type GroupChatTermination struct {
	Reason string
}
