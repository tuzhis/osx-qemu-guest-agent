package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-ping",
		Handler: handleGuestPing,
		Enabled: true,
	})
}

// handleGuestPing handles the guest-ping command.
// It's a no-op command that simply returns an empty object to confirm the
// guest agent is alive and responding.
func handleGuestPing(req json.RawMessage) (interface{}, error) {
	return protocol.EmptyResponse{}, nil
}
