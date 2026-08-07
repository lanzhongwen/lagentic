package agentchat

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_Values(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"MaxTurnsExceeded", ErrMaxTurnsExceeded, "max turns exceeded"},
		{"TokenLimitExceeded", ErrTokenLimitExceeded, "token limit exceeded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("error = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestSentinelErrors_Wrapped(t *testing.T) {
	wrapped := fmt.Errorf("team stopped: %w", ErrMaxTurnsExceeded)
	if !errors.Is(wrapped, ErrMaxTurnsExceeded) {
		t.Error("errors.Is should find ErrMaxTurnsExceeded through wrap")
	}
}
