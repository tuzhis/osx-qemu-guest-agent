package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-time",
		Handler: handleGetTime,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-set-time",
		Handler: handleSetTime,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-get-timezone",
		Handler: handleGetTimezone,
		Enabled: true,
	})
}

// handleGetTime handles the guest-get-time command.
func handleGetTime(req json.RawMessage) (interface{}, error) {
	now := time.Now()
	// Return nanoseconds timestamp.
	nanoseconds := now.UnixNano()

	logrus.WithFields(logrus.Fields{
		"timestamp": nanoseconds,
		"time":      now.Format(time.RFC3339),
	}).Info("Getting system time")

	return nanoseconds, nil
}

// handleSetTime handles the guest-set-time command.
func handleSetTime(req json.RawMessage) (interface{}, error) {
	var args protocol.SetTimeArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, err
	}

	targetTime := time.Unix(0, args.Time)

	// Setting system time on macOS requires administrator privileges.
	if err := setSystemTime(targetTime); err != nil {
		logrus.WithError(err).Error("Failed to set system time")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"time":      targetTime.Format(time.RFC3339),
		"timestamp": targetTime.UnixNano(),
	}).Info("System time set successfully")

	return protocol.EmptyResponse{}, nil
}

// handleGetTimezone handles the guest-get-timezone command.
func handleGetTimezone(req json.RawMessage) (interface{}, error) {
	now := time.Now()
	zone, offset := now.Zone()

	result := &protocol.GuestTimezone{
		Zone:   zone,
		Offset: offset,
	}

	logrus.WithFields(logrus.Fields{
		"zone":   zone,
		"offset": offset,
	}).Info("Getting timezone information")

	return result, nil
}

// setSystemTime sets the system time.
func setSystemTime(t time.Time) error {
	// On macOS, use the date command to set the time.
	// Format: MMddHHmmYY (MonthDayHourMinuteYear)
	dateStr := t.Format("0102150406")
	cmd := exec.Command("sudo", "date", dateStr)
	return cmd.Run()
}
