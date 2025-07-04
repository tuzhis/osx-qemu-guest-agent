package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-suspend-disk",
		Handler: handleGuestSuspendDisk,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-suspend-ram",
		Handler: handleGuestSuspendRAM,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-suspend-hybrid",
		Handler: handleGuestSuspendHybrid,
		Enabled: true,
	})
}

// handleGuestSuspendDisk handles the guest-suspend-disk command.
func handleGuestSuspendDisk(req json.RawMessage) (interface{}, error) {
	logrus.Info("Executing guest-suspend-disk command")
	if err := suspendToDisk(); err != nil {
		logrus.WithError(err).Error("Suspend to disk failed")
		return nil, err
	}
	logrus.Info("guest-suspend-disk command executed successfully")
	return protocol.EmptyResponse{}, nil
}

// handleGuestSuspendRAM handles the guest-suspend-ram command.
func handleGuestSuspendRAM(req json.RawMessage) (interface{}, error) {
	logrus.Info("Executing guest-suspend-ram command")
	if err := suspendToRAM(); err != nil {
		logrus.WithError(err).Error("Suspend to RAM failed")
		return nil, err
	}
	logrus.Info("guest-suspend-ram command executed successfully")
	return protocol.EmptyResponse{}, nil
}

// handleGuestSuspendHybrid handles the guest-suspend-hybrid command.
func handleGuestSuspendHybrid(req json.RawMessage) (interface{}, error) {
	logrus.Info("Executing guest-suspend-hybrid command")
	if err := suspendHybrid(); err != nil {
		logrus.WithError(err).Error("Hybrid suspend failed")
		return nil, err
	}
	logrus.Info("guest-suspend-hybrid command executed successfully")
	return protocol.EmptyResponse{}, nil
}

// suspendToDisk suspends the system to disk.
func suspendToDisk() error {
	// On macOS, use the pmset command for power management.
	// hibernatemode 25 means suspend to disk.
	cmd := exec.Command("pmset", "-a", "hibernatemode", "25")
	if err := cmd.Run(); err != nil {
		return err
	}
	// Execute the sleep command.
	cmd = exec.Command("pmset", "sleepnow")
	return cmd.Run()
}

// suspendToRAM suspends the system to RAM.
func suspendToRAM() error {
	// On macOS, use the pmset command for sleep.
	// hibernatemode 0 means suspend to RAM only.
	cmd := exec.Command("pmset", "-a", "hibernatemode", "0")
	if err := cmd.Run(); err != nil {
		return err
	}
	// Execute the sleep command.
	cmd = exec.Command("pmset", "sleepnow")
	return cmd.Run()
}

// suspendHybrid performs a hybrid suspend.
func suspendHybrid() error {
	// On macOS, use the pmset command for hybrid sleep.
	// hibernatemode 3 means RAM + disk hybrid mode.
	cmd := exec.Command("pmset", "-a", "hibernatemode", "3")
	if err := cmd.Run(); err != nil {
		return err
	}
	// Execute the sleep command.
	cmd = exec.Command("pmset", "sleepnow")
	return cmd.Run()
}
