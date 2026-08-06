package core

import "sync"

// CancellationToken propagates cancellation from user (Ctrl+C) through all layers.
type CancellationToken struct {
	done chan struct{}
	once sync.Once
}

// NewCancellationToken creates an un-cancelled token.
func NewCancellationToken() *CancellationToken {
	return &CancellationToken{
		done: make(chan struct{}),
	}
}

// Cancel closes the done channel. Safe to call multiple times.
func (ct *CancellationToken) Cancel() {
	ct.once.Do(func() { close(ct.done) })
}

// Done returns a channel that is closed when Cancel is called.
func (ct *CancellationToken) Done() <-chan struct{} {
	return ct.done
}

// MessageContext carries per-message metadata.
type MessageContext struct {
	IsRPC             bool
	TopicID           TopicID
	Sender            AgentID
	CancellationToken *CancellationToken
}
