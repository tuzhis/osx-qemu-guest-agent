package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"sort"

	"github.com/sirupsen/logrus"
)

// AgentVersion should be updated with a proper versioning system,
// perhaps using build flags.
const AgentVersion = "1.1.0"

func init() {
	RegisterCommand(&Command{
		Name:    "guest-info",
		Handler: handleGuestInfo,
		Enabled: true,
	})
}

// handleGuestInfo handles the guest-info command, returning information
// about the guest agent and its supported commands.
func handleGuestInfo(req json.RawMessage) (interface{}, error) {
	supportedCommands := getSupportedCommands()
	agentInfo := &protocol.GuestAgentInfo{
		Version:           AgentVersion,
		SupportedCommands: supportedCommands,
	}

	logrus.WithFields(logrus.Fields{
		"version":       agentInfo.Version,
		"command_count": len(supportedCommands),
	}).Info("Returning Guest Agent information")

	return agentInfo, nil
}

// getSupportedCommands retrieves the list of all supported commands directly
// from the CommandRegistry.
func getSupportedCommands() []protocol.GuestAgentCommandInfo {
	// Ensure the list is sorted for consistent output.
	commandNames := make([]string, 0, len(CommandRegistry))
	for name := range CommandRegistry {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	commands := make([]protocol.GuestAgentCommandInfo, 0, len(commandNames))
	for _, name := range commandNames {
		cmd := CommandRegistry[name]
		commandInfo := protocol.GuestAgentCommandInfo{
			Name:    cmd.Name,
			Enabled: cmd.Enabled,
			// QEMU guest agent expects this field, defaulting to true.
			SuccessResponse: true,
		}
		commands = append(commands, commandInfo)
	}

	return commands
}
