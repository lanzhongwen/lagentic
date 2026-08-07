package agentchat

import "fmt"

var (
	ErrMaxTurnsExceeded  = fmt.Errorf("max turns exceeded")
	ErrTokenLimitExceeded = fmt.Errorf("token limit exceeded")
)
