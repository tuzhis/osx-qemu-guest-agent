package commands

import (
	"encoding/json"
	"fmt"
)

// GuestSyncArgs defines the arguments for the guest-sync command.
type GuestSyncArgs struct {
	ID int64 `json:"id"`
}

func init() {
	// The 'guest-sync-id' command was found in the original command map.
	// The official command is 'guest-sync'. I will register both for compatibility.
	RegisterCommand(&Command{
		Name:    "guest-sync",
		Handler: handleGuestSync,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-sync-id",
		Handler: handleGuestSync,
		Enabled: true,
	})
	// guest-sync-delimited uses the same logic as guest-sync. The agent's
	// response handling layer is responsible for sending the delimiter.
	RegisterCommand(&Command{
		Name:    "guest-sync-delimited",
		Handler: handleGuestSync,
		Enabled: true,
	})
}

// handleGuestSync handles both guest-sync and guest-sync-delimited commands.
// It simply returns the ID that was passed in.
func handleGuestSync(req json.RawMessage) (interface{}, error) {
	var args GuestSyncArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments for guest-sync: %v", err)
	}

	return args.ID, nil
}
