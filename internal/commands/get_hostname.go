package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-hostname",
		Handler: handleGetHostname,
		Enabled: true,
	})
	// Register the hyphenated version for backward compatibility with some clients.
	RegisterCommand(&Command{
		Name:    "guest-get-host-name",
		Handler: handleGetHostname,
		Enabled: true,
	})
}

// handleGetHostname handles the guest-get-hostname command.
func handleGetHostname(req json.RawMessage) (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.WithError(err).Error("Failed to get hostname")
		return nil, err
	}

	logrus.WithField("hostname", hostname).Info("Successfully retrieved hostname")

	return protocol.GuestHostName{HostName: hostname}, nil
}
